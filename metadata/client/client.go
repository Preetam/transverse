// Package client contains the transverse metadata service client.
package client

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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	http    *http.Client
	base    string
	token   string
	headers map[string]string
}

type ServerError int

func (e ServerError) Error() string {
	return fmt.Sprintf("client: server status code %d", e)
}

func New(baseURI, token string) *Client {
	return &Client{
		http: &http.Client{
			Timeout: time.Second * 30,
		},
		headers: map[string]string{
			"Accept":       "application/json",
			"Content-Type": "application/json",
		},
		base:  strings.TrimRight(baseURI, "/"),
		token: token,
	}
}

func (c *Client) doRequest(verb string, address string, body, response interface{}) error {
	payload := bytes.NewBuffer(nil)
	if body != nil {
		err := json.NewEncoder(payload).Encode(body)
		if err != nil {
			return err
		}
	}

	request, err := http.NewRequest(verb, c.base+address, payload)
	if err != nil {
		return err
	}

	if c.token != "" {
		request.Header.Set("X-Api-Key", c.token)
	}

	for key, val := range c.headers {
		request.Header.Set(key, val)
	}

	res, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	defer io.Copy(ioutil.Discard, res.Body)

	if res.StatusCode/100 != 2 {
		if response != nil {
			json.NewDecoder(res.Body).Decode(response)
			// Ignore errors
		}
		return ServerError(res.StatusCode)
	}

	if response != nil {
		err := json.NewDecoder(res.Body).Decode(response)
		if err != nil {
			return err
		}
	}

	return nil
}
