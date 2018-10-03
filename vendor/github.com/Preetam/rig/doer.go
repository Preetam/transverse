package rig

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Preetam/lm2log"
	"github.com/Preetam/siesta"
)

type doer struct {
	lock       sync.Mutex
	commitLog  *rigLog
	peer       *logClient
	peerInSync bool
	token      string

	errCount int
}

func newDoer(commitLog *rigLog, peer, token string) (*doer, error) {
	d := &doer{
		commitLog: commitLog,
		token:     token,
	}

	if peer != "" {
		d.peer = newLogClient(peer, token)

		peerCommitted, err := d.peer.Committed()
		if err != nil {
			if err != lm2log.ErrNotFound {
				goto SKIP_PEER
			}
		}
		peerCommittedVersion := peerCommitted.Version

		localCommitted, err := commitLog.Committed()
		if err != nil {
			if err.(logError).StatusCode != http.StatusNotFound {
				return nil, err
			}
		}
		localCommittedVersion := localCommitted.Version

		// Check if peer is behind or caught up (special case).
		if peerCommittedVersion <= localCommittedVersion {
			// It is not. If it is, the loop below does nothing.
			for i := peerCommittedVersion; i != localCommittedVersion; i++ {
				// Get the ith record.
				payload, err := commitLog.Record(i + 1)
				if err != nil {
					log.Println(err)
					goto SKIP_PEER
				}
				err = d.peer.Prepare(payload)
				if err != nil {
					log.Println(err)
					goto SKIP_PEER
				}
				err = d.peer.Commit()
				if err != nil {
					log.Println(err)
					goto SKIP_PEER
				}
			}
		} else {
			// Peer is ahead.
			err = commitLog.Rollback()
			if err != nil {
				return nil, err
			}
			for i := localCommittedVersion; i != peerCommittedVersion; i++ {
				// Get the ith record.
				payload, err := d.peer.GetRecord(i + 1)
				if err != nil {
					return nil, err
				}
				err = commitLog.Prepare(payload)
				if err != nil {
					return nil, err
				}
				err = commitLog.Commit()
				if err != nil {
					return nil, err
				}
			}
		}

		// Now the committed versions are synced up. It's time to handle the prepared case.

		peerPrepared, err := d.peer.Prepared()
		if err != nil {
			if err != lm2log.ErrNotFound {
				goto SKIP_PEER
			}
		}
		peerPreparedVersion := peerPrepared.Version

		localPrepared, err := commitLog.Prepared()
		if err != nil {
			if err.(logError).StatusCode != http.StatusNotFound {
				return nil, err
			}
		}
		localPreparedVersion := localPrepared.Version

		if localPreparedVersion > 0 || peerPreparedVersion > 0 {
			// Something was prepared and not completed.
			// Roll them back.
			err = commitLog.Rollback()
			if err != nil {
				return nil, err
			}
			err = d.peer.Rollback()
			if err != nil {
				goto SKIP_PEER
			}
		}
	}
	d.peerInSync = true
SKIP_PEER:
	err := commitLog.Commit()
	if err != nil {
		return nil, err
	}
	go d.syncPeer()
	return d, nil
}

func (d *doer) do(p LogPayload, ignoreVersion bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	err := d.commitLog.LockResources(p.Op)
	if err != nil {
		return err
	}
	defer d.commitLog.UnlockResources(p.Op)

	committedPayload, err := d.commitLog.Committed()
	if err != nil {
		if err.(logError).Err != nil {
			log.Println("couldn't get prepared version:", err)
			return err
		}
	}

	if ignoreVersion {
		p.Version = committedPayload.Version + 1
	}

	err = d.commitLog.Prepare(p)
	if err != nil {
		log.Println("couldn't prepare locally:", err)
		return err
	}

	if d.peer != nil && d.peerInSync {
		for tries := 0; tries < 3; tries++ {
			err = d.peer.Prepare(p)
			if err != nil {
				log.Println("couldn't prepare on peer:", err)
				continue
			}
			break
		}
		if err != nil {
			log.Println("couldn't prepare on peer:", err)
			log.Println("marking peer as out-of-sync and continuing")
			d.peerInSync = false
		}
	}

	err = d.commitLog.Commit()
	if err != nil {
		log.Println("couldn't commit locally:", err)
		log.Println("rolling back")
		rollbackErr := d.commitLog.Rollback()
		if rollbackErr != nil {
			log.Fatalln("rollback failed:", rollbackErr)
		}
		return err
	}

	if d.peer != nil && d.peerInSync {
		for tries := 0; tries < 3; tries++ {
			err = d.peer.Commit()
			if err != nil {
				log.Println("couldn't commit on peer:", err)
				continue
			}
			break
		}

		if err != nil {
			log.Println("couldn't commit on peer:", err)
			log.Println("marking peer as out-of-sync and continuing")
			d.peerInSync = false
		}
	}

	return nil
}

func (d *doer) Handler() http.Handler {
	service := siesta.NewService("/")
	service.AddPre(requestIdentifier)
	service.AddPre(checkAuth(d.token))
	service.AddPost(responseGenerator)
	service.AddPost(responseWriter)

	service.Route("POST", "/do", "do endpoint", func(c siesta.Context, w http.ResponseWriter, r *http.Request) {
		requestData := c.Get(requestDataKey).(*requestData)

		var params siesta.Params
		ignoreVersion := params.Bool("ignore-version", true, "Ignore version in payload")
		err := params.Parse(r.Form)
		if err != nil {
			requestData.ResponseError = err.Error()
			requestData.StatusCode = http.StatusBadRequest
			return
		}

		var doPayload LogPayload
		err = json.NewDecoder(r.Body).Decode(&doPayload)
		if err != nil {
			requestData.ResponseError = err.Error()
			requestData.StatusCode = http.StatusBadRequest
			return
		}

		err = d.do(doPayload, *ignoreVersion)
		if err != nil {
			requestData.ResponseError = err.Error()
			if logErr, ok := err.(logError); ok {
				requestData.StatusCode = logErr.StatusCode
			} else {
				requestData.StatusCode = http.StatusInternalServerError
			}
			log.Printf("[Req %s] error %v", requestData.RequestID, err)
			return
		}
	})

	return service
}

func (d *doer) syncPeer() {
	if d.peer == nil {
		return
	}
	log.Println("initializing sync")
	sleepDur := 3 * time.Second
	sleep := false
	for {
		if sleep {
			time.Sleep(sleepDur)
		}
		sleep = false

		d.lock.Lock()
		if d.peerInSync {
			// already in sync
			d.lock.Unlock()
			sleep = true
			continue
		}
		log.Println("PEER: peer is not in sync")
		d.lock.Unlock()

		d.peer.Rollback()
		peerCommitted, err := d.peer.Committed()
		if err != nil {
			if err != lm2log.ErrNotFound {
				sleep = true
				log.Println("PEER:", err)
				continue
			}
		}
		peerCommittedVersion := peerCommitted.Version

		d.lock.Lock()
		localCommitted, err := d.commitLog.Committed()
		if err != nil {
			sleep = true
			d.lock.Unlock()
			log.Println("PEER:", err)
			continue
		}
		localCommittedVersion := localCommitted.Version

		if localCommittedVersion == peerCommittedVersion {
			// in sync
			d.peerInSync = true
			d.lock.Unlock()
			log.Println("PEER:", "IN SYNC")
			continue
		}
		d.lock.Unlock()

		for i := peerCommittedVersion; i != localCommittedVersion; i++ {
			// Get the ith record.
			payload, err := d.commitLog.Record(i + 1)
			if err != nil {
				sleep = true
				log.Println("PEER:", err)
				continue
			}
			err = d.peer.Prepare(payload)
			if err != nil {
				sleep = true
				log.Println("PEER:", err)
				continue
			}
			err = d.peer.Commit()
			if err != nil {
				sleep = true
				log.Println("PEER:", err)
				continue
			}
		}
	}
}
