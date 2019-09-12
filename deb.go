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

// Package debpack packs files to deb files.
//
// It is designed to be simple to use and deploy, not requiring any
// filesystem access to create deb files.
// It is influenced heavily in style and interface from the
// github.com/google/rpmpack package.
package debpack

import (
	"bytes"
	"io"
	"sort"
	"text/template"

	"github.com/blakesmith/ar"
	"github.com/pkg/errors"
)

// DEBMetaData contains meta info about the whole package.
type DEBMetaData struct {
	Name          string
	Version       string
	Arch          string
	Maintainer    string
	Description   string
	InstalledSize uint
}

// DEBFile contains a particular file's entry and data.
type DEBFile struct {
	Name  string
	Body  []byte
	Mode  int64
	Owner string
	Group string
	MTime int64
}

// DEB holds the state of a particular deb file. Please use NewDEB to instantiate it.
type DEB struct {
	DEBMetaData
	files map[string]DEBFile
}

// NewDEB creates and returns a new DEB struct.
func NewDEB(m DEBMetaData) (*DEB, error) {
	return &DEB{
		DEBMetaData: m,
		files:       make(map[string]DEBFile),
	}, nil
}

// Write closes the deb and writes the whole deb to an io.Writer
func (d *DEB) Write(w io.Writer) error {
	// Add all of the files, sorted alphabetically.
	fnames := []string{}
	for fn := range d.files {
		fnames = append(fnames, fn)
	}
	sort.Strings(fnames)

	// gen data.tar.gz
	data := newArchive()

	for _, fn := range fnames {
		if err := data.write(d.files[fn]); err != nil {
			return errors.Wrapf(err, "cannot add %q file to data archive", fn)
		}
	}
	md5sums := data.md5sums()
	d.InstalledSize = data.payloadSize / 1024

	if err := data.Close(); err != nil {
		return errors.Wrap(err, "failed to close data archive payload")
	}

	// gen control.tar.gz
	control := newArchive()

	var c bytes.Buffer
	if err := controlTmpl.Execute(&c, d); err != nil {
		return errors.Wrap(err, "failed execute control template")
	}

	if err := control.writeFile("control", c.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "cannot add control file to control archive")
	}

	if err := control.writeFile("md5sums", []byte(md5sums), 0644); err != nil {
		return errors.Wrap(err, "cannot add md5sums file to control archive")
	}

	if err := control.Close(); err != nil {
		return errors.Wrap(err, "failed to close control archive payload")
	}

	a := ar.NewWriter(w)

	if err := a.WriteGlobalHeader(); err != nil {
		return errors.Wrap(err, "failed write ar header")
	}

	if err := writeFile(a, "debian-binary", []byte("2.0\n"), 0644); err != nil {
		return errors.Wrap(err, "cannot add debian-binary to deb")
	}

	if err := writeFile(a, "control.tar.gz", control.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "cannot add control.tar.gz to deb")
	}

	if err := writeFile(a, "data.tar.gz", data.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "cannot add data.tar.gz to deb")
	}

	return nil
}

// AddFile adds an DEBFile to an existing deb.
func (d *DEB) AddFile(f DEBFile) {
	if f.Name == "/" { // deb does not allow the root dir to be included.
		return
	}
	d.files[f.Name] = f
}

func writeFile(a *ar.Writer, name string, data []byte, mode int64) error {
	var header = ar.Header{
		Name: name,
		Size: int64(len(data)),
		Mode: mode,
	}
	if err := a.WriteHeader(&header); err != nil {
		return errors.Wrap(err, "cannot write ar file header")
	}
	_, err := a.Write(data)
	return err
}

var controlTmpl = template.Must(template.New("control").Parse(`
{{- /* Mandatory fields */ -}}
Package: {{.Name}}
Version: {{.Version}}
Architecture: {{.Arch}}
Installed-Size: {{.InstalledSize}}
Maintainer: {{.Maintainer}}
Description: {{.Description}}
`))
