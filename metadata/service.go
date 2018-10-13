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
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/Preetam/lm2"
	"github.com/Preetam/rig"
	"github.com/Preetam/siesta"
	"github.com/Preetam/transverse/metadata/client"
	"github.com/Preetam/transverse/metadata/middleware"
	log "github.com/Sirupsen/logrus"
)

var (
	errNotFound = errors.New("lm2: not found")
)

const (
	prefixUser      = "00:" // users
	prefixUserEmail = "01:" // index for user.Email => user.ID
	prefixGoal      = "02:" // goals
	prefixUserGoal  = "03:" // index for user.ID + goal.ID => ""
	prefixMetadata  = "zz:" // metadata stuff
)

const (
	tupleSeparator = "\x00\x00"
)

type MetadataService struct {
	// Main metadata collection
	col  *lm2.Collection
	lock sync.RWMutex

	dataDir string

	riggedService *rig.RiggedService
}

func NewMetadataService(dataDir string) (*MetadataService, error) {
	collectionPath := filepath.Join(dataDir, "data")
	err := os.MkdirAll(filepath.Join(collectionPath), 0755)
	if err != nil {
		return nil, err
	}
	col, err := lm2.NewCollection(filepath.Join(collectionPath, "data.lm2"), 100000)
	if err != nil {
		return nil, err
	}
	wb := lm2.NewWriteBatch()
	wb.Set(prefixMetadata+"version", "0")
	_, err = col.Update(wb)
	if err != nil {
		col.Destroy()
		return nil, err
	}
	return &MetadataService{
		col:     col,
		dataDir: dataDir,
	}, nil
}

func OpenMetadataService(dataDir string) (*MetadataService, error) {
	collectionPath := filepath.Join(dataDir, "data")
	col, err := lm2.OpenCollection(filepath.Join(collectionPath, "data.lm2"), 100000)
	if err != nil {
		return nil, err
	}
	return &MetadataService{
		col:     col,
		dataDir: dataDir,
	}, nil
}

func (s *MetadataService) Validate(o rig.Operation) error {
	log.Println("Validate", o.Method, string(o.Data))
	switch o.Method {
	case client.OpUserCreate:
		return s.CreateUserValidate(o.Data)
	case client.OpUserUpdate:
		return s.UpdateUserValidate(o.Data)
	case client.OpUserDelete:
		return s.DeleteUserValidate(o.Data)

	case client.OpGoalCreate:
		return s.CreateGoalValidate(o.Data)
	case client.OpGoalUpdate:
		return s.UpdateGoalValidate(o.Data)
	}
	return errors.New("invalid method")
}

func (s *MetadataService) LockResources(o rig.Operation) bool {
	// No-op for now
	return true
}

func (s *MetadataService) UnlockResources(o rig.Operation) {
	// No-op for now
}

func (s *MetadataService) Apply(version uint64, o rig.Operation) error {
	log.Println("Apply", version, o.Method, string(o.Data))
	switch o.Method {
	case client.OpUserCreate:
		return s.CreateUserApply(version, o.Data)
	case client.OpUserUpdate:
		return s.UpdateUserApply(version, o.Data)
	case client.OpUserDelete:
		return s.DeleteUserApply(version, o.Data)

	case client.OpGoalCreate:
		return s.CreateGoalApply(version, o.Data)
	case client.OpGoalUpdate:
		return s.UpdateGoalApply(version, o.Data)

	}
	return errors.New("invalid method")
}

func (s *MetadataService) Version() (uint64, error) {
	cur, err := s.col.NewCursor()
	if err != nil {
		return 0, err
	}
	versionStr, err := cursorGet(cur, prefixMetadata+"version")
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(versionStr, 10, 64)
}

type kvPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (s *MetadataService) Snapshot() (io.ReadSeeker, int64, error) {
	cur, err := s.col.NewCursor()
	if err != nil {
		return nil, 0, err
	}

	snapshotData := []kvPair{}
	for cur.Next() {
		snapshotData = append(snapshotData, kvPair{Key: cur.Key(), Value: cur.Value()})
	}
	marshaled, err := json.Marshal(snapshotData)
	if err != nil {
		return nil, 0, err
	}

	return bytes.NewReader(marshaled), int64(len(marshaled)), nil
}

func (s *MetadataService) Restore(version uint64, r io.Reader) error {
	snapshotData := []kvPair{}
	err := json.NewDecoder(r).Decode(&snapshotData)
	if err != nil {
		return err
	}
	err = s.col.Destroy()
	if err != nil {
		return err
	}
	collectionPath := filepath.Join(s.dataDir, "data")
	s.col, err = lm2.NewCollection(filepath.Join(collectionPath, "data.lm2"), 100000)
	if err != nil {
		return err
	}
	wb := lm2.NewWriteBatch()
	for _, kv := range snapshotData {
		wb.Set(kv.Key, kv.Value)
	}
	_, err = s.col.Update(wb)
	return err
}

func (s *MetadataService) Service() *siesta.Service {
	MetadataService := siesta.NewService("/")
	MetadataService.AddPre(middleware.RequestIdentifier)
	MetadataService.AddPre(middleware.CheckAuth)
	MetadataService.AddPost(middleware.ResponseGenerator)
	MetadataService.AddPost(middleware.ResponseWriter)

	MetadataService.Route("POST", "/do", "Do is the write endpoint", func(c siesta.Context, w http.ResponseWriter, r *http.Request) {
		requestData := c.Get(middleware.RequestDataKey).(*middleware.RequestData)

		var doPayload rig.Operation
		err := json.NewDecoder(r.Body).Decode(&doPayload)
		if err != nil {
			requestData.ResponseError = err.Error()
			requestData.StatusCode = http.StatusBadRequest
			return
		}

		err = s.riggedService.Apply(doPayload, false)
		if err != nil {
			requestData.ResponseError = err.Error()
			requestData.StatusCode = http.StatusInternalServerError
			return
		}
	})

	// Read endpoints

	MetadataService.Route("GET", "/goals/:id", "Gets a goal by ID", s.GetGoal)

	MetadataService.Route("GET", "/users/:id", "Gets a user by ID", s.GetUserByID)
	MetadataService.Route("GET", "/users/:id/goals", "Gets a user's goals", s.GetUserGoals)
	MetadataService.Route("GET", "/users", "Searches for a user", s.GetUsers)

	return MetadataService
}

func cursorGet(cur *lm2.Cursor, key string) (string, error) {
	cur.Seek(key)
	for cur.Next() {
		if cur.Key() > key {
			break
		}
		if cur.Key() == key {
			return cur.Value(), nil
		}
	}
	return "", errNotFound
}
