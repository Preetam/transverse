package main

/**
 * Copyright (C) 2018 Preetam Jinka
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/Preetam/siesta"
	"github.com/Preetam/transverse/internal/forecast1"
	"github.com/Preetam/transverse/internal/forecast3"
	"github.com/Preetam/transverse/metadata/client"
	"github.com/Preetam/transverse/metadata/middleware"
	"github.com/Preetam/transverse/metadata/token"
	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const APIBasePath = "/api/v1/"
const UserContextKey = "user"

type API struct {
	os ObjectStore
}

func NewAPI(os ObjectStore) *API {
	return &API{
		os: os,
	}
}

// Service returns a siesta service for the API.
func (api *API) Service() *siesta.Service {
	APIService := siesta.NewService(APIBasePath)
	APIService.AddPre(middleware.RequestIdentifier)
	APIService.AddPre(api.CheckAuth)
	APIService.AddPost(middleware.ResponseGenerator)
	APIService.AddPost(middleware.ResponseWriter)

	APIService.Route("GET", "/user", "serves user API endpoint", api.GetUser)
	APIService.Route("GET", "/user/data", "gets all user data", api.GetUserData)
	APIService.Route("DELETE", "/user", "serves user deletion API endpoint", api.DeleteUser)
	APIService.Route("PUT", "/user/password", "updates user password", api.PutPassword)
	APIService.Route("GET", "/goals", "serves goals API endpoint", api.GetGoals)
	APIService.Route("GET", "/goals/:goalID", "serves goal API endpoint", api.GetGoal)
	APIService.Route("PUT", "/goals/:goalID", "serves update goal API endpoint", api.UpdateGoal)
	APIService.Route("POST", "/goals", "serves create goal API endpoint", api.CreateGoal)
	APIService.Route("DELETE", "/goals/:goalID", "serves delete goal API endpoint", api.DeleteGoal)

	// Goal data
	APIService.Route("GET", "/goals/:goalID/data", "serves get goal data API endpoint", api.GetGoalData)
	APIService.Route("GET", "/goals/:goalID/eta", "serves get goal eta API endpoint", api.GetGoalETA)
	APIService.Route("GET", "/goals/:goalID/raw-data", "serves get raw goal data API endpoint", api.GetRawGoalData)
	APIService.Route("POST", "/goals/:goalID/data", "serves add goal data API endpoint", api.PostGoalData)
	APIService.Route("POST", "/goals/:goalID/data/single", "serves add goal single data API endpoint", api.PostGoalDataSingle)

	return APIService
}

func (api *API) CheckAuth(c siesta.Context, w http.ResponseWriter, r *http.Request, q func()) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)

	cookie, err := r.Cookie("transverse")
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusUnauthorized
		q()
		return
	}

	userTokenData := &token.UserTokenData{}
	_, err = TokenCodec.DecodeToken(cookie.Value, userTokenData)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusUnauthorized
		q()
		return
	}

	user, err := MetadataClient.GetUserByID(userTokenData.User)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		q()
		return
	}

	if !user.Verified {
		log.Println(requestData.RequestID, "unverified")
		requestData.StatusCode = http.StatusUnauthorized
		q()
		return
	}

	if user.Deleted > 0 {
		log.Println(requestData.RequestID, "deleted")
		requestData.StatusCode = http.StatusUnauthorized
		q()
		return
	}

	c.Set(UserContextKey, userTokenData)
}

func (api *API) GetUser(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)
	userTokenData := c.Get(UserContextKey).(*token.UserTokenData)

	user, err := MetadataClient.GetUserByID(userTokenData.User)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		if serverErr, ok := err.(client.ServerError); ok {
			requestData.StatusCode = int(serverErr)
		}
		return
	}
	user.PasswordHash = ""
	requestData.ResponseData = user
}

func (api *API) GetUserData(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)
	userTokenData := c.Get(UserContextKey).(*token.UserTokenData)

	user, err := MetadataClient.GetUserByID(userTokenData.User)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		if serverErr, ok := err.(client.ServerError); ok {
			requestData.StatusCode = int(serverErr)
		}
		return
	}
	user.PasswordHash = ""

	response := map[string]interface{}{}
	response["user"] = user

	goals, err := MetadataClient.GetUserGoals(user.ID, true)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		if serverErr, ok := err.(client.ServerError); ok {
			requestData.StatusCode = int(serverErr)
		}
		return
	}

	response["goals"] = goals

	goalDataMap := map[string]interface{}{}

	for _, goal := range goals {
		reader, err := api.os.GetObject(goal.ID)
		if err != nil {
			if err == errDoesNotExist {
				continue
			}
			requestData.StatusCode = http.StatusInternalServerError
			return
		}

		goalData := []goalDataPoint{}

		err = json.NewDecoder(reader).Decode(&goalData)
		if err != nil {
			continue
		}

		goalDataMap[goal.ID] = goalData
	}

	response["goal_data"] = goalDataMap
	requestData.ResponseData = response
}

func (api *API) DeleteUser(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)
	userTokenData := c.Get(UserContextKey).(*token.UserTokenData)

	user, err := MetadataClient.GetUserByID(userTokenData.User)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		if serverErr, ok := err.(client.ServerError); ok {
			requestData.StatusCode = int(serverErr)
		}
		return
	}

	err = MetadataClient.DeleteUser(user)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		return
	}
}

func (api *API) PutPassword(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)
	userTokenData := c.Get(UserContextKey).(*token.UserTokenData)

	password := ""
	err := json.NewDecoder(r.Body).Decode(&password)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	if len(password) < 8 {
		log.Println(requestData.RequestID, "short password")
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	user, err := MetadataClient.GetUserByID(userTokenData.User)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		if serverErr, ok := err.(client.ServerError); ok {
			requestData.StatusCode = int(serverErr)
		}
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(requestData.RequestID, "short password")
		requestData.StatusCode = http.StatusInternalServerError
		return
	}

	user.PasswordHash = string(hash)
	err = MetadataClient.UpdateUser(user)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		return
	}
}

func (api *API) GetGoals(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)
	userTokenData := c.Get(UserContextKey).(*token.UserTokenData)

	var params siesta.Params
	showArchived := params.Bool("showArchived", false, "Show archived")
	err := params.Parse(r.Form)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	goals, err := MetadataClient.GetUserGoals(userTokenData.User, *showArchived)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		if serverErr, ok := err.(client.ServerError); ok {
			requestData.StatusCode = int(serverErr)
		}
		return
	}
	requestData.ResponseData = goals
}

func (api *API) GetGoal(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)
	userTokenData := c.Get(UserContextKey).(*token.UserTokenData)

	var params siesta.Params
	goalID := params.String("goalID", "", "Goal ID")
	err := params.Parse(r.Form)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	goal, err := MetadataClient.GetGoal(*goalID)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		if serverErr, ok := err.(client.ServerError); ok {
			requestData.StatusCode = int(serverErr)
		}
		return
	}

	if goal.User != userTokenData.User {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		return
	}

	requestData.ResponseData = goal
}

func (api *API) CreateGoal(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)
	userTokenData := c.Get(UserContextKey).(*token.UserTokenData)

	goal := client.Goal{}
	err := json.NewDecoder(r.Body).Decode(&goal)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	if goal.Name == "" {
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	goal.ID = generateCode(8)
	goal.User = userTokenData.User
	goal.Created = time.Now().Unix()
	goal.Updated = time.Now().Unix()

	err = MetadataClient.CreateGoal(goal)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		if serverErr, ok := err.(client.ServerError); ok {
			requestData.StatusCode = int(serverErr)
		}
		return
	}
	requestData.ResponseData = goal
}

func (api *API) UpdateGoal(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)
	userTokenData := c.Get(UserContextKey).(*token.UserTokenData)

	var params siesta.Params
	goalID := params.String("goalID", "", "Goal ID")
	err := params.Parse(r.Form)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	goal := client.Goal{}
	err = json.NewDecoder(r.Body).Decode(&goal)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	if goal.Name == "" {
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	goal.ID = *goalID
	goal.User = userTokenData.User
	goal.Updated = time.Now().Unix()

	err = MetadataClient.UpdateGoal(goal)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		if serverErr, ok := err.(client.ServerError); ok {
			requestData.StatusCode = int(serverErr)
		}
		return
	}
	requestData.ResponseData = goal
}

func (api *API) DeleteGoal(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)
	userTokenData := c.Get(UserContextKey).(*token.UserTokenData)

	var params siesta.Params
	goalID := params.String("goalID", "", "Goal ID")
	err := params.Parse(r.Form)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	goal, err := MetadataClient.GetGoal(*goalID)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		if serverErr, ok := err.(client.ServerError); ok {
			requestData.StatusCode = int(serverErr)
		}
		return
	}

	if goal.User != userTokenData.User {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusForbidden
		return
	}

	goal.Updated = time.Now().Unix()
	goal.Deleted = time.Now().Unix()

	err = MetadataClient.UpdateGoal(goal)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		if serverErr, ok := err.(client.ServerError); ok {
			requestData.StatusCode = int(serverErr)
		}
		return
	}

	api.os.DeleteObject(*goalID)
}

func (api *API) GetRawGoalData(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)
	userTokenData := c.Get(UserContextKey).(*token.UserTokenData)

	var params siesta.Params
	goalID := params.String("goalID", "", "Goal ID")
	err := params.Parse(r.Form)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	goal, err := MetadataClient.GetGoal(*goalID)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		if serverErr, ok := err.(client.ServerError); ok {
			requestData.StatusCode = int(serverErr)
		}
		return
	}

	if goal.User != userTokenData.User {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusForbidden
		return
	}

	reader, err := api.os.GetObject(*goalID)
	if err != nil {
		if err == errDoesNotExist {
			requestData.StatusCode = http.StatusNotFound
			return
		}
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		return
	}

	goalData := []goalDataPoint{}

	err = json.NewDecoder(reader).Decode(&goalData)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	requestData.ResponseData = map[string]interface{}{
		"series": goalData,
	}
}

func getGoalDataInternal(goal client.Goal, goalData []goalDataPoint) map[string]interface{} {
	type forecastT struct {
		Timestamp time.Time `json:"ts"`
		Value     float64   `json:"value"`
	}

	predictionForecast := []forecastT{}
	lowForecast := []forecastT{}
	highForecast := []forecastT{}

	goalData = fillGaps(goalData)
	haveETA := false
	eta := 0

	if true {
		forecaster := forecast3.NewForecaster()

		lastValue := 0.0
		points := []float64{}
		for _, v := range goalData {
			points = append(points, v.Value)
			lastValue = v.Value
		}
		forecastPoints := forecaster.Forecast(points)
		firstPrediction := 0.0
		for i, v := range forecastPoints {
			if lastValue == goal.Target {
				// Reached goal
				break
			}
			if i == 0 {
				firstPrediction = v.Predicted
			}

			timestamp := goalData[len(goalData)-1].Timestamp.Add(time.Hour * 24 * time.Duration(i))
			predictionForecast = append(predictionForecast, forecastT{
				Timestamp: timestamp,
				Value:     v.Predicted,
			})
			lowForecast = append(lowForecast, forecastT{
				Timestamp: timestamp,
				Value:     v.Low,
			})
			highForecast = append(highForecast, forecastT{
				Timestamp: timestamp,
				Value:     v.High,
			})

			if i > 1 {
				// ETA detection

				if firstPrediction >= goal.Target && predictionForecast[len(predictionForecast)-1].Value < goal.Target {
					haveETA = true
					eta = i - 1
					break
				}

				if firstPrediction <= goal.Target && predictionForecast[len(predictionForecast)-1].Value > goal.Target {
					haveETA = true
					eta = i - 1
					break
				}
			}
		}
	} else {
		// set up initial data
		initialLevel := 0.0
		initialTrend := 0.0
		maxInitialPoints := 4
		if len(goalData) == 1 {
			initialLevel = goalData[0].Value
		} else {
			if len(goalData) < maxInitialPoints {
				maxInitialPoints = len(goalData)
			}

			initialVals := []float64{}
			for i := 0; i < maxInitialPoints; i++ {
				initialVals = append(initialVals, goalData[i].Value)
			}

			initialLevel = sum(initialVals) / float64(maxInitialPoints)
			differences := diff(initialVals)
			initialTrend = sum(differences) / float64(len(differences))
		}

		models := []*forecast1.Model{}
		maxModels := 20
		if len(goalData) < maxModels {
			maxModels = len(goalData)
		}
		for i := 1; i <= maxModels; i++ {
			models = append(models, forecast1.NewModel(i))
		}

		lastValue := 0.0
		for _, point := range goalData {
			models[0].Update(float32(point.Value), float32(initialLevel), float32(initialTrend))
			for j, m := range models {
				if j == 0 {
					continue
				}
				m.Update(float32(point.Value), models[0].Level, models[0].Trend)
			}
			lastValue = point.Value
		}

		firstPrediction := 0.0
		shift := 0.0
		lastSqrtErr := float32(0.0)
		for i, m := range models {
			if lastValue == goal.Target {
				// Reached?
				break
			}
			timestamp := goalData[len(goalData)-1].Timestamp.Add(time.Hour * 24 * time.Duration(i))
			prediction := m.Forecast()
			if i == 0 {
				firstPrediction = float64(prediction)
				shift = lastValue - firstPrediction
			}
			sqrtErr := float32(math.Sqrt(float64(m.ErrEWMA)))
			if sqrtErr < lastSqrtErr {
				sqrtErr = lastSqrtErr
			}
			lastSqrtErr = sqrtErr
			low := prediction - 2*sqrtErr
			high := prediction + 2*sqrtErr

			if i > 1 {
				if firstPrediction >= goal.Target && predictionForecast[len(predictionForecast)-1].Value < goal.Target {
					haveETA = true
					eta = i - 1
					break
				}

				if firstPrediction <= goal.Target && predictionForecast[len(predictionForecast)-1].Value > goal.Target {
					haveETA = true
					eta = i - 1
					break
				}
			}

			predictionForecast = append(predictionForecast, forecastT{
				Timestamp: timestamp,
				Value:     float64(prediction) + shift,
			})

			if i == 0 {
				lowForecast = append(lowForecast, forecastT{
					Timestamp: timestamp,
					Value:     float64(prediction) + shift,
				})
				highForecast = append(highForecast, forecastT{
					Timestamp: timestamp,
					Value:     float64(prediction) + shift,
				})

				continue
			}

			lowForecast = append(lowForecast, forecastT{
				Timestamp: timestamp,
				Value:     float64(low) + shift,
			})
			highForecast = append(highForecast, forecastT{
				Timestamp: timestamp,
				Value:     float64(high) + shift,
			})
		}
	}

	resp := map[string]interface{}{
		"series":     goalData,
		"prediction": predictionForecast,
		"low":        lowForecast,
		"high":       highForecast,
	}
	if haveETA {
		resp["eta"] = eta
	}
	return resp
}

func (api *API) GetGoalData(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)
	userTokenData := c.Get(UserContextKey).(*token.UserTokenData)

	var params siesta.Params
	goalID := params.String("goalID", "", "Goal ID")
	err := params.Parse(r.Form)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	goal, err := MetadataClient.GetGoal(*goalID)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		if serverErr, ok := err.(client.ServerError); ok {
			requestData.StatusCode = int(serverErr)
		}
		return
	}

	if goal.User != userTokenData.User {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusForbidden
		return
	}

	getObjectStartTime := time.Now()
	reader, err := api.os.GetObject(*goalID)
	if err != nil {
		if err == errDoesNotExist {
			requestData.StatusCode = http.StatusNotFound
			return
		}
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		return
	}
	getObjectEndTime := time.Now()
	getObjectLatencyMs := getObjectEndTime.Sub(getObjectStartTime).Seconds() * 1000
	log.WithFields(map[string]interface{}{
		"get_object_latency_ms": getObjectLatencyMs,
	}).Printf("Getting object took %0.2f ms", getObjectLatencyMs)

	goalData := []goalDataPoint{}

	err = json.NewDecoder(reader).Decode(&goalData)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}
	requestData.ResponseData = getGoalDataInternal(goal, goalData)
}

func (api *API) GetGoalETA(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)
	userTokenData := c.Get(UserContextKey).(*token.UserTokenData)

	var params siesta.Params
	goalID := params.String("goalID", "", "Goal ID")
	err := params.Parse(r.Form)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	goal, err := MetadataClient.GetGoal(*goalID)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		if serverErr, ok := err.(client.ServerError); ok {
			requestData.StatusCode = int(serverErr)
		}
		return
	}

	if goal.User != userTokenData.User {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusForbidden
		return
	}

	hasETA := goal.ETA != 0
	if hasETA {
		resp := map[string]interface{}{}
		if goal.ETA == -1 {
			resp["eta"] = nil
		} else {
			resp["eta"] = goal.ETA
		}
		requestData.ResponseData = resp
		return
	}

	getObjectStartTime := time.Now()
	reader, err := api.os.GetObject(*goalID)
	if err != nil {
		if err == errDoesNotExist {
			requestData.StatusCode = http.StatusNotFound
			return
		}
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		return
	}
	getObjectEndTime := time.Now()
	getObjectLatencyMs := getObjectEndTime.Sub(getObjectStartTime).Seconds() * 1000
	log.WithFields(map[string]interface{}{
		"get_object_latency_ms": getObjectLatencyMs,
	}).Printf("Getting object took %0.2f ms", getObjectLatencyMs)

	goalData := []goalDataPoint{}

	err = json.NewDecoder(reader).Decode(&goalData)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}
	resp := map[string]interface{}{}
	resp["eta"] = getGoalDataInternal(goal, goalData)["eta"]
	requestData.ResponseData = resp

	if resp["eta"] != nil {
		goal.ETA = int64(resp["eta"].(int))
	} else {
		goal.ETA = -1
	}
	MetadataClient.UpdateGoal(goal)
}

type goalDataPoint struct {
	Timestamp time.Time `json:"ts"`
	Value     float64   `json:"value"`
}

type pointsByTime []goalDataPoint

func (p pointsByTime) Len() int {
	return len(p)
}

func (p pointsByTime) Less(i, j int) bool {
	return p[i].Timestamp.Before(p[j].Timestamp)
}

func (p pointsByTime) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (api *API) PostGoalData(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)
	userTokenData := c.Get(UserContextKey).(*token.UserTokenData)

	var params siesta.Params
	goalID := params.String("goalID", "", "Goal ID")
	err := params.Parse(r.Form)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	goal, err := MetadataClient.GetGoal(*goalID)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		if serverErr, ok := err.(client.ServerError); ok {
			requestData.StatusCode = int(serverErr)
		}
		return
	}

	if goal.User != userTokenData.User {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusForbidden
		return
	}

	goalData := []goalDataPoint{}
	err = json.NewDecoder(r.Body).Decode(&goalData)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}
	// Format as UTC
	for i, p := range goalData {
		p.Timestamp = p.Timestamp.UTC()
		goalData[i] = p
	}
	sort.Sort(pointsByTime(goalData))

	// marshal it for S3
	marshaled, err := json.Marshal(goalData)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		return
	}
	err = api.os.PutObject(*goalID, bytes.NewReader(marshaled), int64(len(marshaled)))
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		return
	}

	goal.Updated = time.Now().Unix()
	MetadataClient.UpdateGoal(goal)
}

func (api *API) PostGoalDataSingle(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)
	userTokenData := c.Get(UserContextKey).(*token.UserTokenData)

	var params siesta.Params
	goalID := params.String("goalID", "", "Goal ID")
	add := params.Bool("add", false, "Add value; set otherwise")
	err := params.Parse(r.Form)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	goal, err := MetadataClient.GetGoal(*goalID)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		if serverErr, ok := err.(client.ServerError); ok {
			requestData.StatusCode = int(serverErr)
		}
		return
	}

	if goal.User != userTokenData.User {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusForbidden
		return
	}

	type pointPayload struct {
		Date  time.Time `json:"date"`
		Value float64   `json:"value"`
	}
	point := pointPayload{}
	err = json.NewDecoder(r.Body).Decode(&point)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	reader, err := api.os.GetObject(*goalID)
	if err != nil {
		if err != errDoesNotExist {
			log.Println(requestData.RequestID, err)
			requestData.StatusCode = http.StatusInternalServerError
			return
		} else {
			// Maybe it's the first point.
		}
	}

	goalData := []goalDataPoint{}

	if err != errDoesNotExist {
		err = json.NewDecoder(reader).Decode(&goalData)
		if err != nil {
			log.Println(requestData.RequestID, err)
			requestData.StatusCode = http.StatusBadRequest
			return
		}
	}

	hasPoint := false
	today := point.Date.Truncate(24 * time.Hour).UTC()
	for _, p := range goalData {
		if p.Timestamp.Unix() == today.Unix() {
			hasPoint = true
			break
		}
	}
	if !hasPoint {
		goalData = append(goalData, goalDataPoint{
			Timestamp: today,
			Value:     0,
		})
	}
	sort.Sort(pointsByTime(goalData))

	for i, p := range goalData {
		if p.Timestamp.Unix() == today.Unix() {
			newValue := point.Value
			if *add {
				newValue += p.Value
			}
			p.Value = newValue
			goalData[i] = p
			break
		}
	}

	// marshal it for S3
	marshaled, err := json.Marshal(goalData)
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		return
	}
	err = api.os.PutObject(*goalID, bytes.NewReader(marshaled), int64(len(marshaled)))
	if err != nil {
		log.Println(requestData.RequestID, err)
		requestData.StatusCode = http.StatusInternalServerError
		return
	}

	goal.Updated = time.Now().Unix()
	MetadataClient.UpdateGoal(goal)
}

func diff(vals []float64) []float64 {
	differences := []float64{}
	for i := 1; i < len(vals); i++ {
		differences = append(differences, vals[i]-vals[i-1])
	}
	return differences
}

func sum(vals []float64) float64 {
	result := 0.0
	for _, v := range vals {
		result += v
	}
	return result
}

func unixDay(t time.Time) int {
	return int(t.Unix() / int64(86400))
}

func fillGaps(p []goalDataPoint) []goalDataPoint {
	points := []goalDataPoint{}
	current := 0
	end := p[len(p)-1].Timestamp
	for now := p[0].Timestamp; unixDay(now) <= unixDay(end); now = now.Add(24 * time.Hour) {
		if unixDay(p[current].Timestamp) == unixDay(now) {
			points = append(points, p[current])
			current++
			continue
		}
		weight := float64(unixDay(now)-unixDay(p[current-1].Timestamp)) /
			float64(unixDay(p[current].Timestamp)-unixDay(p[current-1].Timestamp))
		newPoint := goalDataPoint{
			Timestamp: now,
			Value:     p[current-1].Value + weight*(p[current].Value-p[current-1].Value),
		}
		points = append(points, newPoint)
	}
	return points
}
