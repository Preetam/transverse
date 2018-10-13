package rig

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

type testService struct {
	version uint64
}

func (s *testService) Version() (uint64, error) {
	return s.version, nil
}

func (s *testService) Validate(Operation) error {
	return nil
}
func (s *testService) Apply(version uint64, op Operation) error {
	s.version = version
	return nil
}
func (s *testService) Snapshot() (io.ReadSeeker, int64, error) {
	snapshot := []byte(fmt.Sprint(s.version))
	r := bytes.NewReader(snapshot)
	return r, int64(len(snapshot)), nil
}
func (s *testService) Restore(version uint64, r io.Reader) error {
	s.version = version
	return nil
}

type testObjectStore struct{ t *testing.T }

func (o *testObjectStore) GetObject(name string) (io.ReadCloser, error) {
	o.t.Log("getting object", name)
	return nil, nil
}

func (o *testObjectStore) DeleteObject(name string) error {
	o.t.Log("deleting object", name)
	return nil
}

func (o *testObjectStore) PutObject(name string, data io.ReadSeeker, size int64) error {
	o.t.Log("putting object", name, "size", size)
	return nil
}

func Test1(t *testing.T) {
	rs, err := NewRiggedService(&testService{}, NewFileObjectStore("/tmp/bucket"), "my_service")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rs.Recover())
	t.Log(rs.service)
	rs.Apply(Operation{}, false)
	rs.Flush()
	rs.Apply(Operation{}, false)
	rs.Apply(Operation{}, false)
	rs.Apply(Operation{}, false)
	rs.Flush()
	rs.Apply(Operation{}, false)
	rs.Flush()
	rs.Snapshot()
}
