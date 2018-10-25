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
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/Preetam/rig"
	"github.com/Preetam/transverse/metadata/middleware"
)

const (
	OpUserCreate = "user_create"
	OpUserUpdate = "user_update"
	OpUserDelete = "user_delete"
)

type User struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash,omitempty"`
	Verified     bool   `json:"verified"`
	Created      int64  `json:"created"`
	Updated      int64  `json:"updated"`
	Deleted      int64  `json:"deleted"`
	LastEmail    int64  `json:"last_email"`
}

func (c *ServiceClient) CreateUser(user User) error {
	marshaled, err := json.Marshal(user)
	if err != nil {
		return err
	}

	payload := rig.Operation{Method: OpUserCreate, Data: marshaled}
	err = c.client.doRequest("POST", "/do?ignore-version=true", &payload, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *ServiceClient) UpdateUser(user User) error {
	marshaled, err := json.Marshal(user)
	if err != nil {
		return err
	}

	payload := rig.Operation{Method: OpUserUpdate, Data: marshaled}
	err = c.client.doRequest("POST", "/do?ignore-version=true", &payload, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *ServiceClient) DeleteUser(user User) error {
	marshaled, err := json.Marshal(user)
	if err != nil {
		return err
	}
	payload := rig.Operation{Method: OpUserDelete, Data: marshaled}
	err = c.client.doRequest("POST", "/do?ignore-version=true", &payload, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *ServiceClient) GetUserByID(id string) (User, error) {
	user := User{}
	resp := middleware.APIResponse{
		Data: &user,
	}
	err := c.client.doRequest("GET", fmt.Sprintf("/users/%s", id), nil, &resp)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (c *ServiceClient) GetUserByEmail(email string) (User, error) {
	user := User{}
	resp := middleware.APIResponse{
		Data: &user,
	}
	err := c.client.doRequest("GET", fmt.Sprintf("/users?email=%s", url.QueryEscape(email)), nil, &resp)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (c *ServiceClient) GetUserGoals(userID string, archived bool) (map[string]Goal, error) {
	goals := map[string]Goal{}
	resp := middleware.APIResponse{
		Data: &goals,
	}
	err := c.client.doRequest("GET", fmt.Sprintf("/users/%s/goals?showArchived=%v", userID, archived), nil, &resp)
	if err != nil {
		return nil, err
	}
	return goals, nil
}
