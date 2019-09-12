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

package debpack

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type archive struct {
	payload     *bytes.Buffer
	payloadSize uint
	gzPayload   *gzip.Writer
	tar         *tar.Writer
	filedigests []string
}

func newArchive() *archive {
	p := &bytes.Buffer{}
	z := gzip.NewWriter(p)
	t := tar.NewWriter(z)
	return &archive{
		payload:   p,
		gzPayload: z,
		tar:       t,
	}
}

// Bytes returns the bytes of the data archive.
func (a *archive) Bytes() []byte {
	return a.payload.Bytes()
}

func (a *archive) md5sums() string {
	return strings.Join(a.filedigests, "")
}

func (a *archive) Close() error {
	if err := a.tar.Close(); err != nil {
		return errors.Wrap(err, "failed to close tar payload")
	}
	if err := a.gzPayload.Close(); err != nil {
		return errors.Wrap(err, "failed to close gzip payload")
	}
	return nil
}

func (a *archive) writeFile(name string, body []byte, mode int64) error {
	hdr := &tar.Header{
		Name: name,
		Mode: mode,
		Size: int64(len(body)),
	}
	if err := a.tar.WriteHeader(hdr); err != nil {
		return errors.Wrap(err, "failed to write payload file header")
	}
	if _, err := a.tar.Write(body); err != nil {
		return errors.Wrap(err, "failed to write payload file content")
	}
	a.payloadSize += uint(len(body))
	return nil
}

func (a *archive) write(f DEBFile) error {
	name, err := filepath.Rel("/", f.Name)
	if err != nil {
		return errors.Wrap(err, "failed to convert name relative path")
	}

	hdr := &tar.Header{
		Name:    name,
		Mode:    int64(f.Mode),
		Size:    int64(len(f.Body)),
		Uname:   f.Owner,
		Gname:   f.Group,
		ModTime: time.Unix(f.MTime, 0),
	}

	switch {
	case f.Mode&040000 != 0: // directory
		hdr.Typeflag = tar.TypeDir
	case f.Mode&0120000 != 0: //  symlink
		hdr.Typeflag = tar.TypeSymlink
		hdr.Linkname = string(f.Body)
		hdr.Size = 0
		f.Body = nil
	default: // regular file
		hdr.Typeflag = tar.TypeReg
		hdr.Mode = f.Mode | 0100000
		a.filedigests = append(a.filedigests, fmt.Sprintf("%x  %s\n", md5.Sum(f.Body), name))
	}

	if err := a.tar.WriteHeader(hdr); err != nil {
		return errors.Wrap(err, "failed to write payload file header")
	}

	if _, err := a.tar.Write(f.Body); err != nil {
		return errors.Wrap(err, "failed to write payload file content")
	}

	a.payloadSize += uint(len(f.Body))
	return nil
}
