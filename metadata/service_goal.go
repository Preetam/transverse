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
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Preetam/lm2"
	"github.com/Preetam/siesta"

	"github.com/Preetam/transverse/metadata/client"
	"github.com/Preetam/transverse/metadata/middleware"
)

// CreateGoalValidate validates a goal_create operation.
func (s *MetadataService) CreateGoalValidate(data []byte) error {
	goal := client.Goal{}
	err := json.Unmarshal(data, &goal)
	if err != nil {
		return err
	}

	// TODO: validate goal ID, etc.

	// Check if the goal exists.
	cur, err := s.col.NewCursor()
	if err != nil {
		return err
	}

	_, err = cursorGet(cur, prefixGoal+goal.ID)
	if err != errNotFound {
		if err == nil {
			return errors.New("goal exists")
		}
		return err
	}

	// Make sure the user exists
	_, err = cursorGet(cur, prefixUser+goal.User)
	if err == errNotFound {
		return errors.New("unknown user")
	} else if err != nil {
		return err
	}

	return nil
}

// CreateGoalApply applies a goal_create operation.
func (s *MetadataService) CreateGoalApply(version uint64, data []byte) error {
	// Check existing version.
	cur, err := s.col.NewCursor()
	if err != nil {
		return err
	}

	versionStr, err := cursorGet(cur, prefixMetadata+"version")
	if err != nil {
		return err
	}

	existingVersion, err := strconv.ParseUint(versionStr, 10, 64)
	if err != nil {
		return err
	}

	if existingVersion >= version {
		return nil
	}

	goal := client.Goal{}
	err = json.Unmarshal(data, &goal)
	if err != nil {
		return err
	}

	marshaledGoal, err := json.Marshal(goal)
	if err != nil {
		return err
	}

	wb := lm2.NewWriteBatch()
	wb.Set(prefixGoal+goal.ID, string(marshaledGoal))
	wb.Set(prefixUserGoal+goal.User+tupleSeparator+goal.ID, "")
	wb.Set(prefixMetadata+"version", strconv.FormatUint(version, 10))
	_, err = s.col.Update(wb)

	return err
}

// UpdateGoalValidate validates a goal_update operation.
func (s *MetadataService) UpdateGoalValidate(data []byte) error {
	goal := client.Goal{}
	err := json.Unmarshal(data, &goal)
	if err != nil {
		return err
	}

	// TODO: validate goal ID, etc.

	// Check if the goal exists.
	cur, err := s.col.NewCursor()
	if err != nil {
		return err
	}

	_, err = cursorGet(cur, prefixGoal+goal.ID)
	if err == errNotFound {
		return errors.New("goal doesn't exist")
	} else if err != nil {
		return err
	}

	// Make sure the user exists
	_, err = cursorGet(cur, prefixUser+goal.User)
	if err == errNotFound {
		return errors.New("unknown user")
	}

	return nil
}

// UpdateGoalApply applies a goal_update operation.
func (s *MetadataService) UpdateGoalApply(version uint64, data []byte) error {
	// Check existing version.
	cur, err := s.col.NewCursor()
	if err != nil {
		return err
	}

	versionStr, err := cursorGet(cur, prefixMetadata+"version")
	if err != nil {
		return err
	}

	existingVersion, err := strconv.ParseUint(versionStr, 10, 64)
	if err != nil {
		return err
	}

	if existingVersion >= version {
		return nil
	}

	goal := client.Goal{}
	err = json.Unmarshal(data, &goal)
	if err != nil {
		return err
	}

	marshaledGoal, err := json.Marshal(goal)
	if err != nil {
		return err
	}

	wb := lm2.NewWriteBatch()
	wb.Set(prefixGoal+goal.ID, string(marshaledGoal))
	wb.Set(prefixMetadata+"version", strconv.FormatUint(version, 10))
	_, err = s.col.Update(wb)

	return err
}

// GetGoal returns a goal given its ID.
func (s *MetadataService) GetGoal(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)

	var params siesta.Params
	id := params.String("id", "", "")
	err := params.Parse(r.Form)
	if err != nil {
		requestData.ResponseError = err.Error()
		requestData.StatusCode = http.StatusBadRequest
		return
	}

	cur, err := s.col.NewCursor()
	if err != nil {
		requestData.ResponseError = err.Error()
		requestData.StatusCode = http.StatusInternalServerError
		return
	}

	goalStr, err := cursorGet(cur, prefixGoal+*id)
	if err != nil {
		if err == errNotFound {
			requestData.StatusCode = http.StatusNotFound
			return
		}
		requestData.ResponseError = err.Error()
		requestData.StatusCode = http.StatusInternalServerError
		return
	}

	goal := client.Goal{}
	err = json.Unmarshal([]byte(goalStr), &goal)
	if err != nil {
		requestData.ResponseError = err.Error()
		requestData.StatusCode = http.StatusInternalServerError
		return
	}

	requestData.ResponseData = goal
}
