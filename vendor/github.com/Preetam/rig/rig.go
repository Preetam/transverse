package rig

import "net/http"

type Rig struct {
	d         *doer
	commitLog *rigLog

	// auth token
	token string
}

// New returns a new Rig.
func New(logDir string, service Service, applyCommits bool, token, peer string) (*Rig, error) {
	commitLog, err := newRigLog(logDir, token, service, applyCommits)
	if err != nil {
		return nil, err
	}
	d, err := newDoer(commitLog, peer, token)
	if err != nil {
		return nil, err
	}
	return &Rig{
		d:         d,
		commitLog: commitLog,
	}, nil
}

func (r *Rig) LogHandler() http.Handler {
	return r.commitLog.Handler()
}

func (r *Rig) DoHandler() http.Handler {
	return r.d.Handler()
}
