package rig

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/Preetam/lm2log"
	"github.com/Preetam/siesta"
)

type logError struct {
	// Error type
	Type string
	// HTTP status code
	StatusCode int
	// Underlying error, if any.
	Err error
}

func (err logError) Error() string {
	return fmt.Sprintf("logError [%s-%d] (%s)", err.Type, err.StatusCode, err.Err)
}

// HTTPError represents an HTTP error that has an associated status code.
type HTTPError interface {
	StatusCode() int
}

type Service interface {
	Validate(Operation) error
	Apply(uint64, Operation) error
	LockResources(Operation) bool
	UnlockResources(Operation)
}

// rigLog represents a commit log.
type rigLog struct {
	// service is the service being modified.
	service Service
	// apply determines whether or not commits are
	// applied to the service.
	applyCommits bool
	// commitLog represents the actual log on disk.
	commitLog *lm2log.Log

	// token represents a token used for authentication
	token string

	lock sync.Mutex
}

func newRigLog(logDir string, token string, service Service, applyCommits bool) (*rigLog, error) {
	collectionPath := filepath.Join(logDir, "log")
	err := os.MkdirAll(collectionPath, 0755)
	if err != nil {
		return nil, err
	}
	commitLog, err := lm2log.Open(filepath.Join(collectionPath, "log.lm2"))
	if err != nil {
		if err == lm2log.ErrDoesNotExist {
			commitLog, err = lm2log.New(filepath.Join(collectionPath, "log.lm2"))
		}
		if err != nil {
			return nil, logError{Type: "commitlog_new", Err: err}
		}
	}

	return &rigLog{
		service:      service,
		applyCommits: applyCommits,
		commitLog:    commitLog,
		token:        token,
	}, nil
}

func (l *rigLog) Prepared() (LogPayload, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	var p LogPayload

	preparedVersion, err := l.commitLog.Prepared()
	if err != nil {
		if err == lm2log.ErrNotFound {
			return p, logError{
				Type:       "internal",
				Err:        err,
				StatusCode: http.StatusNotFound,
			}
		}
		return p, logError{
			Type:       "internal",
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	preparedData, err := l.commitLog.Get(preparedVersion)
	if err != nil {
		return p, logError{
			Type:       "internal",
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	operation := Operation{}
	err = json.Unmarshal([]byte(preparedData), &operation)
	if err != nil {
		oldOperationStruct := struct {
			Method string          `json:"method"`
			Data   json.RawMessage `json:"data"`
		}{}
		err = json.Unmarshal([]byte(preparedData), &oldOperationStruct)
		if err == nil {
			operation.Method = oldOperationStruct.Method
			operation.Data = []byte(oldOperationStruct.Data)
		} else {
			return p, logError{
				Type:       "internal",
				Err:        err,
				StatusCode: http.StatusInternalServerError,
			}
		}
	}

	p = LogPayload{
		Version: preparedVersion,
		Op:      operation,
	}

	return p, nil
}

func (l *rigLog) Committed() (LogPayload, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	var p LogPayload

	committedVersion, err := l.commitLog.Committed()
	if err != nil {
		return p, logError{
			Type:       "internal",
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	if committedVersion == 0 {
		return p, logError{
			Type:       "internal",
			Err:        nil,
			StatusCode: http.StatusNotFound,
		}
	}

	committedData, err := l.commitLog.Get(committedVersion)
	if err != nil {
		return p, logError{
			Type:       "internal",
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	operation := Operation{}
	err = json.Unmarshal([]byte(committedData), &operation)
	if err != nil {
		oldOperationStruct := struct {
			Method string          `json:"method"`
			Data   json.RawMessage `json:"data"`
		}{}
		err = json.Unmarshal([]byte(committedData), &oldOperationStruct)
		if err == nil {
			operation.Method = oldOperationStruct.Method
			operation.Data = []byte(oldOperationStruct.Data)
		} else {
			return p, logError{
				Type:       "internal",
				Err:        err,
				StatusCode: http.StatusInternalServerError,
			}
		}
	}

	p = LogPayload{
		Version: committedVersion,
		Op:      operation,
	}

	return p, nil
}

func (l *rigLog) Prepare(payload LogPayload) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	committed, err := l.commitLog.Committed()
	if err != nil {
		return logError{
			Type:       "internal",
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	if payload.Version != committed+1 {
		return logError{
			Type:       "internal",
			Err:        errors.New("preparing invalid version"),
			StatusCode: http.StatusBadRequest,
		}
	}

	err = l.service.Validate(payload.Op)
	if err != nil {
		statusCode := http.StatusBadRequest
		if httpErr, ok := err.(HTTPError); ok {
			statusCode = httpErr.StatusCode()
		}
		return logError{
			Type:       "internal",
			Err:        errors.New("invalid operation"),
			StatusCode: statusCode,
		}
	}

	data, err := json.Marshal(payload.Op)
	if err != nil {
		return logError{
			Type:       "internal",
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	err = l.commitLog.Prepare(string(data))
	if err != nil {
		return logError{
			Type:       "internal",
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	return nil
}

func (l *rigLog) Commit() error {
	l.lock.Lock()
	defer l.lock.Unlock()

	err := l.commitLog.Commit()
	if err != nil {
		return logError{
			Type:       "internal",
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	committedVersion, err := l.commitLog.Committed()
	if err != nil {
		return logError{
			Type:       "internal",
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	if committedVersion == 0 {
		return nil
	}

	committedData, err := l.commitLog.Get(committedVersion)
	if err != nil {
		return logError{
			Type:       "internal",
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	operation := Operation{}
	err = json.Unmarshal([]byte(committedData), &operation)
	if err != nil {
		oldOperationStruct := struct {
			Method string          `json:"method"`
			Data   json.RawMessage `json:"data"`
		}{}
		err = json.Unmarshal([]byte(committedData), &oldOperationStruct)
		if err == nil {
			operation.Method = oldOperationStruct.Method
			operation.Data = []byte(oldOperationStruct.Data)
		} else {
			return logError{
				Type:       "internal",
				Err:        err,
				StatusCode: http.StatusInternalServerError,
			}
		}
	}

	err = l.service.Apply(committedVersion, operation)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if httpErr, ok := err.(HTTPError); ok {
			statusCode = httpErr.StatusCode()
		}
		return logError{
			Type:       "internal",
			Err:        err,
			StatusCode: statusCode,
		}
	}

	return nil
}

func (l *rigLog) Rollback() error {
	l.lock.Lock()
	defer l.lock.Unlock()

	err := l.commitLog.Rollback()
	if err != nil {
		return logError{
			Type:       "internal",
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	return nil
}

func (l *rigLog) Record(version uint64) (LogPayload, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	var p LogPayload

	committedData, err := l.commitLog.Get(version)
	if err != nil {
		return p, logError{
			Type:       "internal",
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	operation := Operation{}
	err = json.Unmarshal([]byte(committedData), &operation)
	if err != nil {
		oldOperationStruct := struct {
			Method string          `json:"method"`
			Data   json.RawMessage `json:"data"`
		}{}
		err = json.Unmarshal([]byte(committedData), &oldOperationStruct)
		if err == nil {
			operation.Method = oldOperationStruct.Method
			operation.Data = []byte(oldOperationStruct.Data)
		} else {
			return p, logError{
				Type:       "internal",
				Err:        err,
				StatusCode: http.StatusInternalServerError,
			}
		}
	}

	p = LogPayload{
		Version: version,
		Op:      operation,
	}

	return p, nil
}

func (l *rigLog) LockResources(o Operation) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	locked := l.service.LockResources(o)
	if !locked {
		return logError{
			Type:       "internal",
			Err:        errors.New("resource busy"),
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	return nil
}

func (l *rigLog) UnlockResources(o Operation) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.service.UnlockResources(o)
}

func (l *rigLog) Compact(recordsToKeep uint) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	err := l.commitLog.Compact(recordsToKeep)
	if err != nil {
		return logError{
			Type:       "internal",
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	return nil
}

func (l *rigLog) Handler() http.Handler {
	commitLog := l.commitLog

	logService := siesta.NewService("/")
	logService.AddPre(requestIdentifier)
	logService.AddPre(checkAuth(l.token))
	logService.AddPost(responseGenerator)
	logService.AddPost(responseWriter)

	logService.Route("GET", "/log/prepare", "", func(c siesta.Context, w http.ResponseWriter, r *http.Request) {
		requestData := c.Get(requestDataKey).(*requestData)
		payload, err := l.Prepared()
		if err != nil {
			requestData.ResponseError = err.Error()
			requestData.StatusCode = err.(logError).StatusCode
			return
		}
		requestData.ResponseData = payload
	})

	logService.Route("POST", "/log/prepare", "", func(c siesta.Context, w http.ResponseWriter, r *http.Request) {
		requestData := c.Get(requestDataKey).(*requestData)
		var preparePayload LogPayload

		err := json.NewDecoder(r.Body).Decode(&preparePayload)
		if err != nil {
			requestData.ResponseError = err.Error()
			requestData.StatusCode = http.StatusBadRequest
			return
		}

		err = l.Prepare(preparePayload)
		if err != nil {
			requestData.ResponseError = err.Error()
			requestData.StatusCode = err.(logError).StatusCode
			return
		}
	})

	logService.Route("GET", "/log/commit", "", func(c siesta.Context, w http.ResponseWriter, r *http.Request) {
		requestData := c.Get(requestDataKey).(*requestData)
		payload, err := l.Committed()
		if err != nil {
			requestData.ResponseError = err.Error()
			requestData.StatusCode = err.(logError).StatusCode
			return
		}
		requestData.ResponseData = payload
	})

	logService.Route("POST", "/log/rollback", "", func(c siesta.Context, w http.ResponseWriter, r *http.Request) {
		requestData := c.Get(requestDataKey).(*requestData)
		err := commitLog.Rollback()
		if err != nil {
			requestData.ResponseError = err.Error()
			requestData.StatusCode = http.StatusInternalServerError
			return
		}
	})

	logService.Route("POST", "/log/commit", "", func(c siesta.Context, w http.ResponseWriter, r *http.Request) {
		requestData := c.Get(requestDataKey).(*requestData)
		err := l.Commit()
		if err != nil {
			requestData.ResponseError = err.Error()
			requestData.StatusCode = err.(logError).StatusCode
			return
		}
	})

	logService.Route("GET", "/log/record/:id", "", func(c siesta.Context, w http.ResponseWriter, r *http.Request) {
		requestData := c.Get(requestDataKey).(*requestData)

		var params siesta.Params
		id := params.Uint64("id", 0, "")
		err := params.Parse(r.Form)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		data, err := commitLog.Get(*id)
		if err != nil {
			requestData.ResponseError = err.Error()
			requestData.StatusCode = http.StatusInternalServerError
			return
		}

		operation := Operation{}
		err = json.Unmarshal([]byte(data), &operation)
		if err != nil {
			oldOperationStruct := struct {
				Method string          `json:"method"`
				Data   json.RawMessage `json:"data"`
			}{}
			err = json.Unmarshal([]byte(data), &oldOperationStruct)
			if err == nil {
				operation.Method = oldOperationStruct.Method
				operation.Data = []byte(oldOperationStruct.Data)
			} else {
				requestData.ResponseError = err.Error()
				requestData.StatusCode = http.StatusInternalServerError
				return
			}
		}

		payload := LogPayload{
			Version: *id,
			Op:      operation,
		}

		requestData.ResponseData = payload
	})

	logService.Route("POST", "/log/compact", "", func(c siesta.Context, w http.ResponseWriter, r *http.Request) {
		requestData := c.Get(requestDataKey).(*requestData)
		var params siesta.Params
		keep := params.Uint64("keep", 10000, "Records to keep")
		err := params.Parse(r.Form)
		if err != nil {
			requestData.ResponseError = err.Error()
			requestData.StatusCode = http.StatusBadRequest
			return
		}

		err = l.Compact(uint(*keep))
		if err != nil {
			requestData.ResponseError = err.Error()
			requestData.StatusCode = err.(logError).StatusCode
			return
		}
	})

	return logService
}
