// This is the transverse metadata service.
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
	"flag"
	"net/http"
	"time"

	"github.com/Preetam/lm2"
	"github.com/Preetam/rig"
	"github.com/Preetam/transverse/metadata/middleware"
	log "github.com/Sirupsen/logrus"
)

var buildStr = "[DEV]"
var versionStr = "0.0.0"

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	listenAddr := flag.String("listen", "localhost:4000", "Listen address")
	dataDir := flag.String("data-dir", "/tmp/data", "Data directory")
	flag.StringVar(&middleware.Token, "token", middleware.Token, "Auth token")
	flag.Parse()

	MetadataService, err := OpenMetadataService(*dataDir)
	if err != nil {
		if err == lm2.ErrDoesNotExist {
			MetadataService, err = NewMetadataService(*dataDir)
			if err != nil {
				log.Fatal("couldn't create metadata service:", err)
			}
		} else {
			log.Fatal("couldn't create metadata service:", err)
		}
	}

	riggedService, err := rig.NewRiggedService(MetadataService, rig.NewFileObjectStore(*dataDir), "rig")
	if err != nil {
		log.Fatal(err)
	}

	MetadataService.riggedService = riggedService

	http.Handle("/", MetadataService.Service())
	log.Println("metadata starting...")

	err = riggedService.Recover()
	if err != nil {
		log.Fatal(err)
	}
	if riggedService.SnapshotVersion() == 0 {
		// Don't have a previous snapshot, so take one now.
		if localVersion, err := MetadataService.Version(); err == nil {
			if localVersion > 0 {
				err = riggedService.Snapshot()
				if err != nil {
					log.Fatalln("error creating initial snapshot:", err)
				} else {
					log.Infoln("created initial snapshot", riggedService.SnapshotVersion())
				}
			}
		} else {
			log.Fatal(err)
		}
	} else {
		log.Infoln("Recovered version", riggedService.SnapshotVersion())
	}

	go func() {
		snapshotTimer := time.Tick(time.Minute)
		flushTimer := time.Tick(time.Second)
		for {
			select {
			case <-snapshotTimer:
				err := riggedService.Snapshot()
				if err != nil {
					log.Warnln("error snapshotting:", err)
				} else {
					log.Infoln("successfully snapshotted version", riggedService.SnapshotVersion())
				}
			case <-flushTimer:
				err := riggedService.Flush()
				if err != nil {
					log.Warnln("error flushing:", err)
				}
			}
		}
	}()

	http.ListenAndServe(*listenAddr, nil)
}
