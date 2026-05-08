// This file is part of go-updater.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package releaser

import (
	"testing"
)

func TestParsePlatform_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected Platform
	}{
		{"linux-amd64", Platform{"linux", "amd64"}},
		{"darwin-arm64", Platform{"darwin", "arm64"}},
		{"windows-386", Platform{"windows", "386"}},
	}

	for _, tt := range tests {
		got, err := Parse(tt.input)
		if err != nil {
			t.Errorf("ParsePlatform(%q) unexpected error: %v", tt.input, err)
		}
		if got != tt.expected {
			t.Errorf("ParsePlatform(%q) = %+v, want %+v", tt.input, got, tt.expected)
		}
	}
}

func TestParsePlatform_Invalid(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"Empty", ""},
		{"Missing Arch", "linux"},
		{"Missing OS", "-amd64"},
		{"Extra Parts", "linux-amd64-extra"},
		{"Just Dash", "-"},
	}

	for _, c := range cases {
		_, err := Parse(c.input)
		if err == nil {
			t.Errorf("ParsePlatform(%q) expected error, got nil", c.input)
		}
	}
}

func TestPlatform_String(t *testing.T) {
	p := Platform{"linux", "amd64"}
	if got := p.String(); got != "linux-amd64" {
		t.Errorf("Platform.String() = %q, want %q", got, "linux-amd64")
	}
}

func TestPlatform_Set(t *testing.T) {
	var p Platform
	err := p.Set("darwin-arm64")
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	if p.OS != "darwin" || p.Arch != "arm64" {
		t.Errorf("Set did not set fields correctly: %+v", p)
	}
}

func TestPlatform_Set_Invalid(t *testing.T) {
	var p Platform
	err := p.Set("badformat")
	if err == nil {
		t.Error("Set should return error for bad format")
	}
}
