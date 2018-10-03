package rig

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/Preetam/siesta"
)

const (
	requestIDKey     = "request-id"
	statusCodeKey    = "status-code"
	responseDataKey  = "response-data"
	responseErrorKey = "response-error"
	responseKey      = "response"
	requestDataKey   = "request-data"
)

type requestData struct {
	RequestID     string
	StatusCode    int
	ResponseData  interface{}
	ResponseError string
	Response      interface{}
	Start         time.Time
}

func requestIdentifier(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := &requestData{
		RequestID: fmt.Sprintf("%08x", rand.Intn(0xffffffff)),
		Start:     time.Now(),
	}
	log.Printf("[Req %s] %s %s", requestData.RequestID, r.Method, r.URL)
	c.Set(requestDataKey, requestData)
}

func responseGenerator(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(requestDataKey).(*requestData)
	response := apiResponse{}

	if data := requestData.ResponseData; data != nil {
		response.Data = data
	}

	response.Err = requestData.ResponseError

	if response.Data != nil || response.Err != "" {
		c.Set(responseKey, response)
	}
}

func responseWriter(c siesta.Context, w http.ResponseWriter, r *http.Request,
	quit func()) {
	requestData := c.Get(requestDataKey).(*requestData)
	if requestData.RequestID != "" {
		w.Header().Set("X-Request-Id", requestData.RequestID)
	}

	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)

	if requestData.StatusCode == 0 {
		requestData.StatusCode = 200
	}
	w.WriteHeader(requestData.StatusCode)

	response := c.Get(responseKey)
	if response != nil {
		enc.Encode(response)
	}

	quit()

	log.Printf("[Req %s] status code %d, latency %0.2f ms", requestData.RequestID, requestData.StatusCode,
		time.Now().Sub(requestData.Start).Seconds()*1000)
}

func checkAuth(token string) func(c siesta.Context, w http.ResponseWriter, r *http.Request, q func()) {
	return func(c siesta.Context, w http.ResponseWriter, r *http.Request, q func()) {
		requestData := c.Get(requestDataKey).(*requestData)
		if token == "" {
			// No token defined
			return
		}
		if r.Header.Get("X-Api-Key") != token {
			requestData.StatusCode = http.StatusUnauthorized
			requestData.ResponseError = "invalid token"
			q()
			return
		}
	}
}
