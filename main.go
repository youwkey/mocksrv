// Copyright 2023 youwkey. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// main
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

const (
	defaultMountDir          = "./"
	defaultAllHost           = false
	defaultListenPort        = "3333"
	defaultReadHeaderTimeout = 3
)

type flags struct {
	version bool
	rootDir string
	allHost bool
	port    string
}

type options struct {
	flags

	addr string
}

//nolint:gochecknoglobals
var (
	version  = "unknown"
	fVersion bool
	fRootDir string
	fAllHost bool
	fPort    string
)

//nolint:gochecknoinits
func init() {
	flag.BoolVar(&fVersion, "version", false, "print the current version")
	flag.StringVar(&fRootDir, "dir", defaultMountDir, "mount root directory")
	flag.BoolVar(&fAllHost, "all", defaultAllHost, "if set, bind any host 0.0.0.0")
	flag.StringVar(&fPort, "port", defaultListenPort, "listen port")

	if version == "unknown" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
			version = info.Main.Version

			return
		}
	}
}

func parseOptions() options {
	flag.Parse()

	host := "127.0.0.1"
	if fAllHost {
		host = "0.0.0.0"
	}

	addr := host + ":" + fPort

	return options{
		flags: flags{
			version: fVersion,
			rootDir: fRootDir,
			allHost: fAllHost,
			port:    fPort,
		},
		addr: addr,
	}
}

func buildHandler(rootDir string) http.Handler {
	return http.FileServer(http.Dir(rootDir))
}

func main() {
	opts := parseOptions()

	if opts.version {
		//nolint:forbidigo
		fmt.Printf("version=%s\n", version)

		return
	}

	handler := buildHandler(opts.rootDir)

	//nolint:exhaustruct
	server := &http.Server{
		Addr:              opts.addr,
		Handler:           handler,
		ReadHeaderTimeout: defaultReadHeaderTimeout * time.Second,
	}

	slog.Info("root dir mounted.", "dir", opts.rootDir)
	slog.Info("server started.", "address", "http://localhost:"+opts.port)

	err := server.ListenAndServe()
	if err != nil {
		slog.Error(err.Error())
	}
}
