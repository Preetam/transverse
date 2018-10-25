package rig

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Service interface {
	Version() (uint64, error)
	Validate(Operation) error
	Apply(uint64, Operation) error
	Snapshot() (io.ReadSeeker, int64, error)
	Restore(uint64, io.Reader) error
}

type Operation struct {
	Method string `json:"method"`
	Data   []byte `json:"data"`
}

type RiggedService struct {
	service          Service
	currentVersion   uint64
	prefix           string
	objectStore      ObjectStore
	pending          []Operation
	lastFlush        uint64
	lastSnapshot     uint64
	lastSnapshotTime time.Time
	lock             sync.Mutex
}

func NewRiggedService(service Service, objectStore ObjectStore, prefix string) (*RiggedService, error) {
	if dirObjectStore, ok := objectStore.(DirectoryCreator); ok {
		err := dirObjectStore.CreateDirectory(filepath.Join(prefix, "LOG"))
		if err != nil {
			return nil, err
		}
		err = dirObjectStore.CreateDirectory(filepath.Join(prefix, "SNAPSHOT"))
		if err != nil {
			return nil, err
		}
	}
	currentVersion, err := service.Version()
	if err != nil {
		return nil, err
	}
	return &RiggedService{
		service:        service,
		objectStore:    objectStore,
		prefix:         prefix,
		currentVersion: currentVersion,
	}, nil
}

func (rs *RiggedService) Recover() error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	err := rs.recoverLatestSnapshot()
	if err != nil {
		return err
	}
	rs.lastFlush = rs.currentVersion
	for {
		err = rs.recoverLogBatch(rs.currentVersion + 1)
		if err != nil {
			if err == errDoesNotExist {
				return nil
			}
			return err
		}
		rs.lastFlush = rs.currentVersion
	}
}

func (rs *RiggedService) recoverLatestSnapshot() error {
	r, err := rs.objectStore.GetObject(rs.getLatestObjectName())
	if err != nil {
		if err == errDoesNotExist {
			return nil
		}
		return err
	}
	defer r.Close()
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	snapshotVersion, err := strconv.ParseUint(string(b), 16, 64)
	if err != nil {
		return err
	}
	sr, err := rs.objectStore.GetObject(rs.getSnapshotName(snapshotVersion))
	if err != nil {
		return err
	}
	defer sr.Close()
	err = rs.service.Restore(snapshotVersion, sr)
	if err != nil {
		return err
	}
	rs.currentVersion = snapshotVersion
	rs.lastSnapshot = snapshotVersion
	return nil
}

func (rs *RiggedService) recoverLogBatch(version uint64) error {
	r, err := rs.objectStore.GetObject(rs.getLogRecordName(version))
	if err != nil {
		return err
	}
	defer r.Close()
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	pending := []Operation{}
	err = json.NewDecoder(gzipReader).Decode(&pending)
	if err != nil {
		return err
	}
	for _, op := range pending {
		err = rs.service.Apply(version, op)
		if err != nil {
			return err
		}
		rs.currentVersion = version
		version++
	}
	return nil
}

// Apply applies an operation.
func (rs *RiggedService) Apply(op Operation, waitUntilDurable bool) error {
	err := rs.service.Validate(op)
	if err != nil {
		return err
	}
	rs.lock.Lock()
	err = rs.service.Apply(rs.currentVersion+1, op)
	if err != nil {
		rs.lock.Unlock()
		return err
	}
	rs.currentVersion++
	rs.pending = append(rs.pending, op)
	rs.lock.Unlock()
	if !waitUntilDurable {
		return nil
	}

	waitFlushVersion := atomic.LoadUint64(&rs.currentVersion)

	timeout := time.NewTimer(10 * time.Second)
	checkTimer := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-timeout.C:
			timeout.Stop()
			checkTimer.Stop()
			return errors.New("rig: timeout")
		case <-checkTimer.C:
			if atomic.LoadUint64(&rs.lastFlush) >= waitFlushVersion {
				timeout.Stop()
				checkTimer.Stop()
				return nil
			}
		}
	}
	// Unreachable
}

func (rs *RiggedService) Flush() (int, error) {
	rs.lock.Lock()
	defer rs.lock.Unlock()

	numRecords := len(rs.pending)
	if numRecords == 0 {
		// Nothing to do
		return 0, nil
	}

	batchVersion := rs.lastFlush + 1

	buf := bytes.NewBuffer(nil)
	w := gzip.NewWriter(buf)
	err := json.NewEncoder(w).Encode(rs.pending)
	if err != nil {
		return 0, err
	}
	w.Flush()
	w.Close()
	err = rs.objectStore.PutObject(rs.getLogRecordName(batchVersion), bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return 0, err
	}
	rs.pending = rs.pending[:0]
	rs.lastFlush = batchVersion + uint64(numRecords) - 1
	return numRecords, nil
}

func (rs *RiggedService) Snapshot() error {
	rs.lock.Lock()
	defer rs.lock.Unlock()

	snapshotVersion, err := rs.service.Version()
	if err != nil {
		return err
	}
	if snapshotVersion == rs.lastSnapshot {
		if time.Now().Before(rs.lastSnapshotTime.Add(24 * time.Hour)) {
			// Less than a day since we took the snapshot, so avoid
			// taking another one. If it's been longer, take it again
			// to be friendly with lifecycle management.
			return nil
		}
	}
	r, size, err := rs.service.Snapshot()
	if err != nil {
		return err
	}
	if closer, ok := r.(io.Closer); ok {
		defer closer.Close()
	}
	err = rs.objectStore.PutObject(rs.getSnapshotName(snapshotVersion), r, size)
	if err != nil {
		return err
	}
	latestFileContents := []byte(strconv.FormatUint(snapshotVersion, 16))
	err = rs.objectStore.PutObject(rs.getLatestObjectName(), bytes.NewReader(latestFileContents), int64(len(latestFileContents)))
	if err != nil {
		return err
	}
	rs.lastSnapshot = snapshotVersion
	rs.lastSnapshotTime = time.Now()
	if rs.currentVersion < snapshotVersion {
		rs.currentVersion = snapshotVersion
	}
	// We won't have any pending records anymore.
	// TODO: keep all records since the last snapshot so we can
	// support point-in-time recovery.
	rs.pending = rs.pending[:0]
	rs.lastFlush = snapshotVersion
	return nil
}

func (rs *RiggedService) SnapshotVersion() uint64 {
	return atomic.LoadUint64(&rs.lastSnapshot)
}

func (rs *RiggedService) getSnapshotName(snapshot uint64) string {
	return filepath.Join(rs.prefix, "SNAPSHOT", fmt.Sprintf("%016x", snapshot))
}

func (rs *RiggedService) getLogRecordName(version uint64) string {
	return filepath.Join(rs.prefix, "LOG", fmt.Sprintf("%016x", version))
}

func (rs *RiggedService) getLatestObjectName() string {
	return filepath.Join(rs.prefix, "LATEST")
}
