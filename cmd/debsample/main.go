// Copyright 2019 Dominik Zeromski <dzeromsk@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// debsample creates an deb file with some known files, which
// can be used to test debpack's output against other deb implementations.
// It is also an instructive example for using debpack.
package main

import (
	"log"
	"os"

	"debpack"
)

func main() {

	r, err := debpack.NewDEB(debpack.DEBMetaData{
		Name:        "debsample",
		Version:     "0.0.1",
		Arch:        "all",
		Maintainer:  "unknown",
		Description: "example package",
	})
	if err != nil {
		log.Fatal(err)
	}
	r.AddFile(
		debpack.DEBFile{
			Name:  "/var/lib/debpack/",
			Mode:  040755,
			Owner: "root",
			Group: "root",
		})
	r.AddFile(
		debpack.DEBFile{
			Name:  "/var/lib/debpack/sample.txt",
			Body:  []byte("testsample\n"),
			Mode:  0600,
			Owner: "root",
			Group: "root",
		})
	r.AddFile(
		debpack.DEBFile{
			Name:  "/var/lib/debpack/sample2.txt",
			Body:  []byte("testsample2\n"),
			Mode:  0644,
			Owner: "root",
			Group: "root",
		})
	r.AddFile(
		debpack.DEBFile{
			Name:  "/var/lib/debpack/sample3_link.txt",
			Body:  []byte("/var/lib/debpack/sample.txt"),
			Mode:  0120777,
			Owner: "root",
			Group: "root",
		})
	if err := r.Write(os.Stdout); err != nil {
		log.Fatalf("write failed: %v", err)
	}

}
