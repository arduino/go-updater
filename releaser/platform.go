// This file is part of go-updater.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of go-updater.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package releaser

import (
	"errors"
	"strings"
)

type Platform struct {
	OS   string
	Arch string
}

func NewPlatform(os string, arch string) Platform {
	return Platform{
		OS:   os,
		Arch: arch,
	}
}

// Parse parses a string like "linux-amd64" into a Platform struct.
func Parse(s string) (Platform, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return Platform{}, errors.New("platform string must be in the form os-arch, e.g. linux-amd64")
	}
	os := parts[0]
	if os == "" {
		return Platform{}, errors.New("missing OS in platform string")
	}
	arch := parts[1]
	if arch == "" {
		return Platform{}, errors.New("missing Arch in platform string")
	}
	return Platform{OS: os, Arch: arch}, nil
}

func MustParse(s string) Platform {
	id, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}

// String returns the platform as "os-arch"
func (p Platform) String() string {
	return p.OS + "-" + p.Arch
}

// Set parses and sets the platform from a string like "linux-amd64"
// Used for flag.Value interface
func (p *Platform) Set(s string) error {
	platform, err := Parse(s)
	if err != nil {
		return err
	}
	p.OS = platform.OS
	p.Arch = platform.Arch
	return nil
}
