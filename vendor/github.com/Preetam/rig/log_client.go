package rig

import (
	"fmt"
	"net/http"

	"github.com/Preetam/lm2log"
)

type logClient struct {
	client *client
}

// LogPayload is request payload for log operations.
type LogPayload struct {
	Version uint64    `json:"version"`
	Op      Operation `json:"op"`
}

func NewLogPayload(version uint64, op Operation) *LogPayload {
	return &LogPayload{
		Version: version,
		Op:      op,
	}
}

// Operation represents a log operation.
type Operation struct {
	Method string `json:"method"`
	Data   []byte `json:"data"`
}

// NewOperation returns a new Operation.
func NewOperation(method string, data []byte) Operation {
	return Operation{
		Method: method,
		Data:   data,
	}
}

func newLogClient(baseURI, token string) *logClient {
	return &logClient{
		client: newClient(baseURI, token),
	}
}

func (c *logClient) Prepared() (LogPayload, error) {
	payload := LogPayload{}
	resp := apiResponse{
		Data: &payload,
	}
	err := c.client.doRequest("GET", "/prepare", nil, &resp)
	if err != nil {
		if serverErr, ok := err.(serverError); ok {
			if serverErr == http.StatusNotFound {
				return payload, lm2log.ErrNotFound
			}
		}
		return payload, resp
	}
	return payload, nil
}

func (c *logClient) Committed() (LogPayload, error) {
	payload := LogPayload{}
	resp := apiResponse{
		Data: &payload,
	}
	err := c.client.doRequest("GET", "/commit", nil, &resp)
	if err != nil {
		if serverErr, ok := err.(serverError); ok {
			if serverErr == http.StatusNotFound {
				return payload, lm2log.ErrNotFound
			}
		}
		return payload, resp
	}
	return payload, nil
}

func (c *logClient) Prepare(payload LogPayload) error {
	err := c.client.doRequest("POST", "/prepare", &payload, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *logClient) Commit() error {
	return c.client.doRequest("POST", "/commit", nil, nil)
}

func (c *logClient) Rollback() error {
	return c.client.doRequest("POST", "/rollback", nil, nil)
}

func (c *logClient) GetRecord(version uint64) (LogPayload, error) {
	p := LogPayload{}
	resp := apiResponse{
		Data: &p,
	}
	err := c.client.doRequest("GET", fmt.Sprintf("/record/%d", version), nil, &resp)
	if err != nil {
		return p, err
	}
	return p, nil
}
