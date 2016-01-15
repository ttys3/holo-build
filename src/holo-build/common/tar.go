/*******************************************************************************
*
* Copyright 2016 Stefan Majewsky <majewsky@gmx.net>
*
* This file is part of Holo.
*
* Holo is free software: you can redistribute it and/or modify it under the
* terms of the GNU General Public License as published by the Free Software
* Foundation, either version 3 of the License, or (at your option) any later
* version.
*
* Holo is distributed in the hope that it will be useful, but WITHOUT ANY
* WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
* A PARTICULAR PURPOSE. See the GNU General Public License for more details.
*
* You should have received a copy of the GNU General Public License along with
* Holo. If not, see <http://www.gnu.org/licenses/>.
*
*******************************************************************************/

package common

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"strings"
	"time"
)

//ToTarArchive creates a TAR archive containing this directory and all the
//filesystem entries in it.
//
//With `leadingDot = true`, generate entry paths like `./foo/bar.conf`.
//WIth `leadingDot = false`, generate entry paths like `foo/bar.conf`.
func (d *FSDirectory) ToTarArchive(leadingDot, buildReproducibly bool) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	var timestamp time.Time
	if buildReproducibly {
		timestamp = time.Unix(0, 0)
	} else {
		timestamp = time.Now()
	}

	err := d.Walk(".", func(path string, node FSNode) error {
		if !leadingDot {
			path = strings.TrimPrefix(path, "./")
		}

		var err error
		switch n := node.(type) {
		case *FSDirectory:
			err = tw.WriteHeader(&tar.Header{
				Name:       path + "/",
				Mode:       040000 | int64(uint(n.Metadata.Mode)&07777),
				Uid:        int(n.Metadata.UID()),
				Gid:        int(n.Metadata.GID()),
				ModTime:    timestamp,
				AccessTime: timestamp,
				ChangeTime: timestamp,
			})
		case *FSRegularFile:
			err = tw.WriteHeader(&tar.Header{
				Name:       path,
				Size:       int64(len([]byte(n.Content))),
				Mode:       0100000 | int64(uint(n.Metadata.Mode)&07777),
				Uid:        int(n.Metadata.UID()),
				Gid:        int(n.Metadata.GID()),
				ModTime:    timestamp,
				AccessTime: timestamp,
				ChangeTime: timestamp,
			})
		case *FSSymlink:
			err = tw.WriteHeader(&tar.Header{
				Name:       path,
				Mode:       0120000 | int64(0777),
				Linkname:   n.Target,
				ModTime:    timestamp,
				AccessTime: timestamp,
				ChangeTime: timestamp,
			})
		default:
			panic("unreachable")
		}
		if err != nil {
			return err
		}
		if n, ok := node.(*FSRegularFile); ok {
			_, err = tw.Write([]byte(n.Content))
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = tw.Close()
	return buf.Bytes(), err
}

//ToTarGZArchive is identical to ToTarArchive, but GZip-compresses the result.
func (d *FSDirectory) ToTarGZArchive(leadingDot, buildReproducibly bool) ([]byte, error) {
	data, err := d.ToTarArchive(leadingDot, buildReproducibly)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)

	_, err = w.Write(data)
	if err != nil {
		return nil, err
	}

	err = w.Close()
	return buf.Bytes(), err
}
