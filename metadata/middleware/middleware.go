package middleware

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
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/Preetam/siesta"
	log "github.com/Sirupsen/logrus"
)

var Token = ""
var VersionStr = ""

const (
	RequestIDKey     = "request-id"
	StatusCodeKey    = "status-code"
	ResponseDataKey  = "response-data"
	ResponseErrorKey = "response-error"
	ResponseKey      = "response"
	RequestDataKey   = "request-data"
)

type RequestData struct {
	RequestID     string
	StatusCode    int
	ResponseData  interface{}
	ResponseError string
	Response      interface{}
	Start         time.Time
}

type APIResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func RequestIdentifier(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := &RequestData{
		RequestID: fmt.Sprintf("%08x", rand.Intn(0xffffffff)),
		Start:     time.Now(),
	}
	log.
		WithField("request_id", requestData.RequestID).
		WithField("method", r.Method).
		WithField("url", r.URL.String()).
		Printf("[Req %s] %s %s", requestData.RequestID, r.Method, r.URL)
	c.Set(RequestDataKey, requestData)
}

func ResponseGenerator(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(RequestDataKey).(*RequestData)
	response := APIResponse{}

	if data := requestData.ResponseData; data != nil {
		response.Data = data
	}

	response.Error = requestData.ResponseError

	if response.Data != nil || response.Error != "" {
		c.Set(ResponseKey, response)
	}
}

func ResponseWriter(c siesta.Context, w http.ResponseWriter, r *http.Request, q func()) {
	requestData := c.Get(RequestDataKey).(*RequestData)
	if requestData.RequestID != "" {
		w.Header().Set("X-Request-Id", requestData.RequestID)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Metadata-Version", VersionStr)

	enc := json.NewEncoder(w)

	if requestData.StatusCode == 0 {
		requestData.StatusCode = 200
	}
	w.WriteHeader(requestData.StatusCode)

	response := c.Get(ResponseKey)
	if response != nil {
		enc.Encode(response)
	}
	q()

	log.Printf("[Req %s] status code %d, latency %0.2f ms", requestData.RequestID, requestData.StatusCode,
		time.Now().Sub(requestData.Start).Seconds()*1000)
	log.
		WithField("request_id", requestData.RequestID).
		WithField("method", r.Method).
		WithField("url", r.URL.String()).
		WithField("status", requestData.StatusCode).
		WithField("latency", time.Now().Sub(requestData.Start).Seconds()*1000).
		Printf("[Req %s] status code %d, latency %0.2f ms", requestData.RequestID, requestData.StatusCode,
			time.Now().Sub(requestData.Start).Seconds()*1000)
}

func CheckAuth(c siesta.Context, w http.ResponseWriter, r *http.Request, q func()) {
	requestData := c.Get(RequestDataKey).(*RequestData)
	if Token == "" {
		// No token defined
		return
	}
	if r.Header.Get("X-Api-Key") != Token {
		requestData.StatusCode = http.StatusUnauthorized
		requestData.ResponseError = "invalid token"
		q()
		return
	}
}
