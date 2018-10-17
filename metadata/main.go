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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var buildStr = "[DEV]"
var versionStr = "0.0.0"

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	listenAddr := flag.String("listen", "localhost:4000", "Listen address")
	dataDir := flag.String("data-dir", "/tmp/data", "Data directory")
	s3Key := flag.String("s3-key", "", "S3 access key")
	s3Secret := flag.String("s3-secret", "", "S3 secret access key")
	s3Region := flag.String("s3-region", "nyc3", "S3 region")
	s3Endpoint := flag.String("s3-endpoint", "https://nyc3.digitaloceanspaces.com", "S3 endpoint")
	flag.StringVar(&middleware.Token, "token", middleware.Token, "Auth token")
	flag.Parse()

	s3Service := s3.New(session.New(aws.NewConfig().WithRegion(*s3Region).WithEndpoint(*s3Endpoint).WithCredentials(credentials.NewStaticCredentials(*s3Key, *s3Secret, ""))))
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

	var objectStore rig.ObjectStore
	if *s3Key == "" {
		objectStore = rig.NewFileObjectStore(*dataDir)
	} else {
		objectStore = rig.NewS3ObjectStore(s3Service, "transverse-rig")
	}
	riggedService, err := rig.NewRiggedService(MetadataService, objectStore, "rig")
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
				start := time.Now()
				err := riggedService.Snapshot()
				if err != nil {
					log.Warnln("error snapshotting:", err)
				} else {
					log.WithField("latency", time.Now().Sub(start).Seconds()).
						Infoln("successfully snapshotted version", riggedService.SnapshotVersion())
				}
			case <-flushTimer:
				start := time.Now()
				flushedCount, err := riggedService.Flush()
				if err != nil {
					log.Warnln("error flushing:", err)
				} else if flushedCount > 0 {
					log.WithField("num_records", flushedCount).
						WithField("latency", time.Now().Sub(start).Seconds()).
						Info("Completed flush")
				}
			}
		}
	}()

	http.ListenAndServe(*listenAddr, nil)
}
