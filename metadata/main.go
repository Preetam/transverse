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
	applyCommits := flag.Bool("apply-commits", true, "Apply log commits")
	peer := flag.String("peer", "", "Peer log base URI")
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

	r, err := rig.New(*dataDir, MetadataService, *applyCommits, middleware.Token, *peer)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/log/", r.LogHandler())
	http.Handle("/", MetadataService.Service())
	http.Handle("/do", r.DoHandler())
	log.Println("metadata starting...")
	http.ListenAndServe(*listenAddr, nil)
}
