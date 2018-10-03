// Package lm2log implements a commit log on top of lm2.
package lm2log

import (
	"errors"
	"strconv"

	"github.com/Preetam/lm2"
)

const (
	preparedKey  = "prepared"
	committedKey = "committed"
)

var (
	ErrNotFound     = errors.New("lm2log: not found")
	ErrDoesNotExist = errors.New("lm2log: does not exist")
)

// Log is a commit log. NOTE: It is *not* goroutine-safe.
type Log struct {
	file string
	col  *lm2.Collection
}

// New initializes a commit log in an lm2 collection.
func New(file string) (*Log, error) {
	col, err := lm2.NewCollection(file, 100)
	if err != nil {
		return nil, err
	}

	l := &Log{
		file: file,
		col:  col,
	}

	wb := lm2.NewWriteBatch()
	wb.Set(committedKey, "0")
	wb.Delete(preparedKey)
	_, err = col.Update(wb)

	return l, err
}

// Open opens an existing commit log in an lm2 collection.
func Open(file string) (*Log, error) {
	col, err := lm2.OpenCollection(file, 100)
	if err != nil {
		if err == lm2.ErrDoesNotExist {
			return nil, ErrDoesNotExist
		}
		return nil, err
	}

	l := &Log{
		file: file,
		col:  col,
	}

	cur, err := col.NewCursor()
	if err != nil {
		return nil, err
	}

	// Check committedKey
	_, err = cursorGet(cur, committedKey)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// Prepare durably prepares data in the log.
func (l *Log) Prepare(data string) error {
	cur, err := l.col.NewCursor()
	if err != nil {
		return err
	}

	// Check if something is already prepared.
	if _, err := cursorGet(cur, preparedKey); err == nil {
		return errors.New("lm2col: already prepared data")
	}

	// Prepare
	committedStr, err := cursorGet(cur, committedKey)
	if err != nil {
		return err
	}

	committed, err := strconv.ParseUint(committedStr, 10, 64)
	if err != nil {
		return err
	}

	wb := lm2.NewWriteBatch()
	nextRecordNum := strconv.FormatUint(committed+1, 10)
	wb.Set(nextRecordNum, data)
	wb.Set(preparedKey, nextRecordNum)
	_, err = l.col.Update(wb)
	return err
}

// Rollback clears any prepared data.
func (l *Log) Rollback() error {
	cur, err := l.col.NewCursor()
	if err != nil {
		return err
	}

	if _, err = cursorGet(cur, preparedKey); err != nil {
		if err == ErrNotFound {
			return nil
		}
		return err
	}

	wb := lm2.NewWriteBatch()
	wb.Delete(preparedKey)
	_, err = l.col.Update(wb)
	return err
}

// Commit marks the prepared data as committed.
func (l *Log) Commit() error {
	cur, err := l.col.NewCursor()
	if err != nil {
		return err
	}

	// Check if something is already prepared.
	prepared, err := cursorGet(cur, preparedKey)
	if err != nil {
		if err == ErrNotFound {
			return nil
		}
		return errors.New("lm2col: couldn't get prepared data")
	}

	// Commit
	wb := lm2.NewWriteBatch()
	wb.Set(committedKey, prepared)
	wb.Delete(preparedKey)
	_, err = l.col.Update(wb)

	return err
}

// Prepared returns the currently prepared record number.
func (l *Log) Prepared() (uint64, error) {
	cur, err := l.col.NewCursor()
	if err != nil {
		return 0, err
	}

	// Check if something is already prepared.
	prepared, err := cursorGet(cur, preparedKey)
	if err != nil {
		if err == ErrNotFound {
			return 0, ErrNotFound
		}
		return 0, errors.New("lm2col: couldn't get prepared data")
	}

	return strconv.ParseUint(prepared, 10, 64)
}

// Committed returns the latest committed record number.
func (l *Log) Committed() (uint64, error) {
	cur, err := l.col.NewCursor()
	if err != nil {
		return 0, err
	}

	committed, err := cursorGet(cur, committedKey)
	if err != nil {
		return 0, errors.New("lm2col: couldn't get committed data")
	}

	return strconv.ParseUint(committed, 10, 64)
}

// Get retrieves data in the commit log associated with the given record number.
func (l *Log) Get(record uint64) (string, error) {
	cur, err := l.col.NewCursor()
	if err != nil {
		return "", err
	}

	// Check if something is already prepared.
	data, err := cursorGet(cur, strconv.FormatUint(record, 10))
	if err != nil {
		return "", errors.New("lm2col: couldn't get record")
	}
	return data, nil
}

// SetCommitted updates the data at the given record number and updates
// any internal state accordingly.
func (l *Log) SetCommitted(record uint64, data string) error {
	cur, err := l.col.NewCursor()
	if err != nil {
		return err
	}

	wb := lm2.NewWriteBatch()
	recordStr := strconv.FormatUint(record, 10)
	committedStr, err := cursorGet(cur, committedKey)
	committed, err := strconv.ParseUint(committedStr, 10, 64)
	if err != nil {
		return err
	}
	if committed < record {
		wb.Set(committedStr, recordStr)
	}

	wb.Set(recordStr, data)
	_, err = l.col.Update(wb)
	return err
}

// Compact compacts the log and keeps up to recordsToKeep records.
func (l *Log) Compact(recordsToKeep uint) error {
	committed, err := l.Committed()
	if err != nil {
		return err
	}
	minRecord := uint64(0)
	if committed > uint64(recordsToKeep) {
		minRecord = committed - uint64(recordsToKeep)
	}
	err = l.col.CompactFunc(func(key string, value string) (string, string, bool) {
		switch key {
		case preparedKey, committedKey:
			return key, value, true
		}

		recordNum, err := strconv.ParseUint(key, 10, 64)
		if err != nil {
			return key, value, true
		}

		if recordNum >= minRecord {
			return key, value, true
		}

		return "", "", false
	})
	if err != nil {
		return err
	}

	col, err := lm2.OpenCollection(l.file, 100)
	if err != nil {
		return err
	}
	l.col = col
	return nil
}

// Close closes the log.
func (l *Log) Close() {
	l.col.Close()
}

// Destroy removes the log.
func (l *Log) Destroy() error {
	return l.col.Destroy()
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
	return "", ErrNotFound
}
