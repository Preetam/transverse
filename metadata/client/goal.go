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

	"github.com/Preetam/rig"
	"github.com/Preetam/transverse/metadata/middleware"
)

type Goal struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	User        string  `json:"user"`
	Description string  `json:"description"`
	Target      float64 `json:"target"`
	ETA         int64   `json:"eta"`
	Archived    bool    `json:"archived"`
	Created     int64   `json:"created"`
	Updated     int64   `json:"updated"`
	Deleted     int64   `json:"deleted"`
}

const (
	OpGoalCreate = "goal_create"
	OpGoalUpdate = "goal_update"
)

func (c *ServiceClient) CreateGoal(goal Goal) error {
	marshaled, err := json.Marshal(goal)
	if err != nil {
		return err
	}

	payload := rig.Operation{Method: OpGoalCreate, Data: marshaled}
	err = c.client.doRequest("POST", "/do?ignore-version=true", &payload, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *ServiceClient) UpdateGoal(goal Goal) error {
	marshaled, err := json.Marshal(goal)
	if err != nil {
		return err
	}

	payload := rig.Operation{Method: OpGoalUpdate, Data: marshaled}
	err = c.client.doRequest("POST", "/do?ignore-version=true", &payload, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *ServiceClient) GetGoal(id string) (Goal, error) {
	goal := Goal{}
	resp := middleware.APIResponse{
		Data: &goal,
	}
	err := c.client.doRequest("GET", fmt.Sprintf("/goals/%s", id), nil, &resp)
	if err != nil {
		return goal, err
	}
	return goal, nil
}
