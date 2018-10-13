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
	"strings"

	"github.com/Preetam/lm2"
	"github.com/Preetam/siesta"
	"github.com/Preetam/transverse/metadata/client"
	"github.com/Preetam/transverse/metadata/middleware"
)

// CreateUserValidate validates a user_create operation.
func (s *MetadataService) CreateUserValidate(data []byte) error {
	user := client.User{}
	err := json.Unmarshal([]byte(data), &user)
	if err != nil {
		return err
	}

	// Check if the user exists.
	cur, err := s.col.NewCursor()
	if err != nil {
		return err
	}

	_, err = cursorGet(cur, prefixUser+user.ID)
	if err != errNotFound {
		if err == nil {
			return errors.New("user exists")
		}
		return err
	}

	_, err = cursorGet(cur, prefixUserEmail+user.Email)
	if err != errNotFound {
		if err == nil {
			return errors.New("user email exists")
		}
		return err
	}

	return nil
}

// CreateUserApply applies a user_create operation.
func (s *MetadataService) CreateUserApply(version uint64, data []byte) error {
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

	user := client.User{}
	err = json.Unmarshal(data, &user)
	if err != nil {
		return err
	}
	marshaledUser, err := json.Marshal(user)
	if err != nil {
		return err
	}

	wb := lm2.NewWriteBatch()
	wb.Set(prefixUser+user.ID, string(marshaledUser))
	wb.Set(prefixUserEmail+user.Email, user.ID)
	wb.Set(prefixMetadata+"version", strconv.FormatUint(version, 10))
	_, err = s.col.Update(wb)

	return err
}

// DeleteUserValidate validates a user_delete operation.
func (s *MetadataService) DeleteUserValidate(data []byte) error {
	user := client.User{}
	err := json.Unmarshal([]byte(data), &user)
	if err != nil {
		return err
	}

	// Check if the user exists.
	cur, err := s.col.NewCursor()
	if err != nil {
		return err
	}

	_, err = cursorGet(cur, prefixUser+user.ID)
	if err == errNotFound {
		return errors.New("user doesn't exist")
	} else if err != nil {
		return err
	}

	userID, err := cursorGet(cur, prefixUserEmail+user.Email)
	if err != nil {
		if err != errNotFound {
			return err
		}
		// errNotFound is OK
	} else {
		if userID != user.ID {
			return errors.New("another user has that email address")
		}
	}

	return nil
}

// DeleteUserApply applies a user_delete operation.
func (s *MetadataService) DeleteUserApply(version uint64, data []byte) error {
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

	user := client.User{}
	err = json.Unmarshal(data, &user)
	if err != nil {
		return err
	}

	existingUserStr, err := cursorGet(cur, prefixUser+user.ID)
	if err != nil {
		return err
	}
	existingUser := client.User{}
	err = json.Unmarshal([]byte(existingUserStr), &existingUser)
	if err != nil {
		return err
	}

	wb := lm2.NewWriteBatch()
	// Delete user
	wb.Delete(prefixUser + user.ID)
	// remove index entries
	wb.Delete(prefixUserEmail + user.Email)

	// Go through goals and delete them
	for cur.Next() {
		if cur.Key() < prefixUserGoal+user.ID+tupleSeparator {
			continue
		}

		if !strings.HasPrefix(cur.Key(), prefixUserGoal+user.ID+tupleSeparator) {
			break
		}
		goalID := strings.TrimPrefix(cur.Key(), prefixUserGoal+user.ID+tupleSeparator)
		wb.Delete(cur.Key())
		wb.Delete(prefixGoal + goalID)
	}

	wb.Set(prefixMetadata+"version", strconv.FormatUint(version, 10))
	_, err = s.col.Update(wb)

	return err
}

// UpdateUserValidate validates a user_update operation.
func (s *MetadataService) UpdateUserValidate(data []byte) error {
	user := client.User{}
	err := json.Unmarshal([]byte(data), &user)
	if err != nil {
		return err
	}

	// Check if the user exists.
	cur, err := s.col.NewCursor()
	if err != nil {
		return err
	}

	_, err = cursorGet(cur, prefixUser+user.ID)
	if err == errNotFound {
		return errors.New("user doesn't exist")
	} else if err != nil {
		return err
	}

	userID, err := cursorGet(cur, prefixUserEmail+user.Email)
	if err != nil {
		if err != errNotFound {
			return err
		}
		// errNotFound is OK
	} else {
		if userID != user.ID {
			return errors.New("another user has that email address")
		}
	}

	return nil
}

// UpdateUserApply applies a user_update operation.
func (s *MetadataService) UpdateUserApply(version uint64, data []byte) error {
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

	user := client.User{}
	err = json.Unmarshal(data, &user)
	if err != nil {
		return err
	}

	existingUserStr, err := cursorGet(cur, prefixUser+user.ID)
	if err != nil {
		return err
	}
	existingUser := client.User{}
	err = json.Unmarshal([]byte(existingUserStr), &existingUser)
	if err != nil {
		return err
	}

	marshaledUser, err := json.Marshal(user)
	if err != nil {
		return err
	}

	wb := lm2.NewWriteBatch()
	wb.Set(prefixUser+user.ID, string(marshaledUser))
	if user.Deleted != 0 {
		// remove index entries
		wb.Delete(prefixUserEmail + user.Email)
		wb.Delete(prefixUserEmail + existingUser.Email)
	} else {
		if user.Email != existingUser.Email {
			// Email address changed, so update index
			wb.Delete(prefixUserEmail + existingUser.Email)
			wb.Set(prefixUserEmail+user.Email, user.ID)
		}
	}
	wb.Set(prefixMetadata+"version", strconv.FormatUint(version, 10))
	_, err = s.col.Update(wb)

	return err
}

func (s *MetadataService) GetUserByID(c siesta.Context, w http.ResponseWriter, r *http.Request) {
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

	userStr, err := cursorGet(cur, prefixUser+*id)
	if err != nil {
		if err == errNotFound {
			requestData.StatusCode = http.StatusNotFound
			return
		}
		requestData.ResponseError = err.Error()
		requestData.StatusCode = http.StatusInternalServerError
		return
	}

	user := client.User{}
	err = json.Unmarshal([]byte(userStr), &user)
	if err != nil {
		requestData.ResponseError = err.Error()
		requestData.StatusCode = http.StatusInternalServerError
		return
	}

	requestData.ResponseData = user
}

func (s *MetadataService) GetUsers(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)

	var params siesta.Params
	email := params.String("email", "", "Email address")
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

	userID, err := cursorGet(cur, prefixUserEmail+*email)
	if err != nil {
		if err == errNotFound {
			requestData.StatusCode = http.StatusNotFound
			return
		}
		requestData.ResponseError = err.Error()
		requestData.StatusCode = http.StatusInternalServerError
		return
	}

	userStr, err := cursorGet(cur, prefixUser+userID)
	if err != nil {
		if err == errNotFound {
			requestData.StatusCode = http.StatusNotFound
			return
		}
		requestData.ResponseError = err.Error()
		requestData.StatusCode = http.StatusInternalServerError
		return
	}

	user := client.User{}
	err = json.Unmarshal([]byte(userStr), &user)
	if err != nil {
		requestData.ResponseError = err.Error()
		requestData.StatusCode = http.StatusInternalServerError
		return
	}

	requestData.ResponseData = user
}

func (s *MetadataService) GetUserGoals(c siesta.Context, w http.ResponseWriter, r *http.Request) {
	requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)

	var params siesta.Params
	id := params.String("id", "", "User ID")
	showArchived := params.Bool("showArchived", false, "Show archived")
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

	goals := map[string]client.Goal{}
	cur.Seek(prefixUserGoal + *id + tupleSeparator)

	goalIDs := []string{}
	for cur.Next() {
		if cur.Key() < prefixUserGoal+*id+tupleSeparator {
			continue
		}

		if !strings.HasPrefix(cur.Key(), prefixUserGoal+*id+tupleSeparator) {
			break
		}
		goalIDs = append(goalIDs,
			strings.TrimPrefix(cur.Key(), prefixUserGoal+*id+tupleSeparator))
	}

	for _, goalID := range goalIDs {
		marshaledGoal, err := cursorGet(cur, prefixGoal+goalID)
		if err != nil {
			requestData.ResponseError = err.Error()
			requestData.StatusCode = http.StatusInternalServerError
			return
		}
		goal := client.Goal{}
		err = json.Unmarshal([]byte(marshaledGoal), &goal)
		if err != nil {
			requestData.ResponseError = err.Error()
			requestData.StatusCode = http.StatusInternalServerError
			return
		}
		if goal.Deleted != 0 {
			// Deleted
			continue
		}
		if !(*showArchived) && goal.Archived {
			// Archived
			continue
		}
		goals[goalID] = goal
	}

	requestData.ResponseData = goals
}
