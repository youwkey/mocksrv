// Copyright 2023 youwkey. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

//nolint:tparallel
func TestParseOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		flags map[string]string
		want  options
	}{
		{name: "without flags", flags: map[string]string{}, want: options{
			flags: flags{version: false, rootDir: defaultMountDir, allHost: defaultAllHost, port: defaultListenPort},
			addr:  "127.0.0.1:" + defaultListenPort,
		}},
		{name: "with dir flag", flags: map[string]string{"dir": "./testdata"}, want: options{
			flags: flags{version: false, rootDir: "./testdata", allHost: defaultAllHost, port: defaultListenPort},
			addr:  "127.0.0.1:" + defaultListenPort,
		}},
		{name: "with all flag", flags: map[string]string{"all": "true"}, want: options{
			flags: flags{version: false, rootDir: defaultMountDir, allHost: true, port: defaultListenPort},
			addr:  "0.0.0.0" + ":" + defaultListenPort,
		}},
		{name: "with port flag", flags: map[string]string{"port": "8080"}, want: options{
			flags: flags{version: false, rootDir: defaultMountDir, allHost: defaultAllHost, port: "8080"},
			addr:  "127.0.0.1:8080",
		}},
	}

	//nolint:paralleltest
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(func() {
				// reset flags
				for k := range test.flags {
					err := flag.CommandLine.Set(k, flag.Lookup(k).DefValue)
					if err != nil {
						t.Fatal(err)
					}
				}
			})

			for k, v := range test.flags {
				err := flag.CommandLine.Set(k, v)
				if err != nil {
					t.Fatal(err)
				}
			}

			got := parseOptions()
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got: %+v want: %+v", got, test.want)
			}
		})
	}
}

//nolint:funlen
func TestBuildHandler(t *testing.T) {
	t.Parallel()

	const rootDir = "./testdata/static"

	handler := buildHandler(rootDir)
	testServer := httptest.NewServer(handler)
	client := new(http.Client)

	t.Cleanup(func() {
		testServer.Close()
	})

	tests := []struct {
		name          string
		filename      string
		status        int
		ignoreContent bool
	}{
		{name: "access html", filename: "index.html", status: 200, ignoreContent: false},
		{name: "access js", filename: "script.js", status: 200, ignoreContent: false},
		{name: "access css", filename: "style.css", status: 200, ignoreContent: false},
		{name: "not found", filename: "not_exits.html", status: 404, ignoreContent: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequestWithContext(
				t.Context(),
				http.MethodGet,
				testServer.URL+"/"+test.filename,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			//nolint:gosec
			res, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() {
				_ = res.Body.Close()
			})

			if res.StatusCode != test.status {
				t.Fatalf("statusCode mismatched: got: %d want %d", res.StatusCode, test.status)
			}

			if test.ignoreContent {
				return
			}

			want, _ := os.ReadFile(filepath.Join(rootDir, test.filename))

			got, _ := io.ReadAll(res.Body)
			if !bytes.Equal(got, want) {
				t.Errorf("content mismatched: %s", test.filename)
			}
		})
	}
}
