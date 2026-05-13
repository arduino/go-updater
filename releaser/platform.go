// This file is part of go-updater.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

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
