// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"testing"
	"unicode/utf8"
)

func TestFomratByteArrayToUtf8(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		want  string
		valid bool
	}{
		{name: "normal", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c}, want: "Hello, 世界", valid: true},
		{name: "substring1", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 0xe7, 0x95}, want: "Hello, 世"},
		{name: "substring2", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 0xe7}, want: "Hello, 世"},
		{name: "substring3", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96}, want: "Hello, 世", valid: true},
		{name: "substring4", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8}, want: "Hello, "},
		{name: "substring5", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4}, want: "Hello, "},
		{name: "substring6", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' '}, want: "Hello, ", valid: true},
		{name: "substring3", data: []byte{'H', 'e', 'l', 'l', 'o', ','}, want: "Hello,", valid: true},
		{name: "invalid", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 'e', 0xe7, 0xe4, 0x95}, want: "Hello, 世e"},
		{name: "invalid2", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 0xe7, 'e', 0xe4, 0x95}, want: "Hello, 世"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatByteArrayToUtf8(tt.data)
			if !utf8.ValidString(got) {
				t.Errorf("Fail to format byteArray to utf8")
			}

			if utf8.Valid(tt.data) != tt.valid {
				t.Errorf("Fail to check validte, got = %v, want %v", !tt.valid, tt.valid)
			}

			if got != tt.want {
				t.Errorf("Fail to check value, got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFomratStringToUtf8(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		want  string
		valid bool
	}{
		{name: "normal", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c}, want: "Hello, 世界", valid: true},
		{name: "substring1", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 0xe7, 0x95}, want: "Hello, 世"},
		{name: "substring2", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 0xe7}, want: "Hello, 世"},
		{name: "substring3", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96}, want: "Hello, 世", valid: true},
		{name: "substring4", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8}, want: "Hello, "},
		{name: "substring5", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4}, want: "Hello, "},
		{name: "substring6", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' '}, want: "Hello, ", valid: true},
		{name: "substring3", data: []byte{'H', 'e', 'l', 'l', 'o', ','}, want: "Hello,", valid: true},
		{name: "invalid", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 'e', 0xe7, 0xe4, 0x95}, want: "Hello, 世e"},
		{name: "invalid2", data: []byte{'H', 'e', 'l', 'l', 'o', ',', ' ', 0xe4, 0xb8, 0x96, 0xe7, 'e', 0xe4, 0x95}, want: "Hello, 世"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := string(tt.data)
			got := FormatStringToUtf8(data)
			if !utf8.ValidString(got) {
				t.Errorf("Fail to format byteArray to utf8")
			}

			if utf8.ValidString(data) != tt.valid {
				t.Errorf("Fail to check validte, got = %v, want %v", !tt.valid, tt.valid)
			}

			if got != tt.want {
				t.Errorf("Fail to check value, got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	datas := []string{
		// Valid
		string("\x45\x00\x45"), // UTF8-1 0x00-0x7F
		string("\x45\x80\x45"), // UTF8-1 Invalid

		string("\x45\xc2\x80\x45"), // UTF8-2 0xC2-0xDF 0x80-0xBF
		string("\x45\xc2\x45"),     // UTF8-2 SubString
		string("\x45\xc2\x7f\x45"), // UTF8-2 Invalid-1

		string("\x45\xe0\xa0\x80\x45"), // UTF8-3 0xE0 0xA0-0xBF 0x80-0xBF
		string("\x45\xe0\xa0\x45"),     // UTF8-3 SubString
		string("\x45\xe0\x45"),         // UTF8-3 SubString
		string("\x45\xe0\xa0\x7f\x45"), // UTF8-3 Invalid-1
		string("\x45\xe0\x9f\x80\x45"), // UTF8-3 Invalid-1
		string("\x45\xe0\x9f\x7f\x45"), // UTF8-3 Invalid-2
		string("\x45\xe0\x9f\x45"),     // UTF8-3 Invalid-1-SubString

		string("\x45\xe1\x80\x80\x45"), // UTF8-3 0xE1-0xEC 0x80-0xBF 0x80-0xBF
		string("\x45\xe1\x80\x45"),     // UTF8-3 SubString
		string("\x45\xe1\x45"),         // UTF8-3 SubString
		string("\x45\xe1\x80\x7f\x45"), // UTF8-3 Invalid-1
		string("\x45\xe1\x7f\x80\x45"), // UTF8-3 Invalid-1
		string("\x45\xe1\x7f\x7f\x45"), // UTF8-3 Invalid-2
		string("\x45\xe1\x7f\x45"),     // UTF8-3 Invalid-1-SubString

		string("\x45\xed\x80\x80\x45"), // UTF8-3 0xED 0x80-0x9F 0x80-0xBF
		string("\x45\xed\x80\x45"),     // UTF8-3 SubString
		string("\x45\xed\x45"),         // UTF8-3 SubString
		string("\x45\xed\x80\x7f\x45"), // UTF8-3 Invalid-1
		string("\x45\xed\x7f\x80\x45"), // UTF8-3 Invalid-1
		string("\x45\xed\x7f\x7f\x45"), // UTF8-3 Invalid-2
		string("\x45\xed\x7f\x45"),     // UTF8-3 Invalid-1-SubString

		string("\x45\xee\x80\x80\x45"), // UTF8-3 0xEE-0xEF 0x80-0xBF 0x80-0xBF
		string("\x45\xee\x80\x45"),     // UTF8-3 SubString
		string("\x45\xee\x45"),         // UTF8-3 SubString
		string("\x45\xee\x80\x7f\x45"), // UTF8-3 Invalid-1
		string("\x45\xee\x7f\x80\x45"), // UTF8-3 Invalid-1
		string("\x45\xee\x7f\x7f\x45"), // UTF8-3 Invalid-2
		string("\x45\xee\x7f\x45"),     // UTF8-3 Invalid-1-SubString

		string("\x45\xf0\x90\x80\x80\x45"), // UTF8-4 0xF0 0x90-0xBF 0x80-0xBF 0x80-0xBF
		string("\x45\xf0\x90\x80\x45"),     // UTF8-4 SubString
		string("\x45\xf0\x90\x45"),         // UTF8-4 SubString
		string("\x45\xf0\x45"),             // UTF8-4 SubString
		string("\x45\xf0\x90\x80\x7f\x45"), // UTF8-4 Invalid-1
		string("\x45\xf0\x90\x7f\x80\x45"), // UTF8-4 Invalid-1
		string("\x45\xf0\x8f\x80\x80\x45"), // UTF8-4 Invalid-1
		string("\x45\xf0\x90\x7f\x7f\x45"), // UTF8-4 Invalid-2
		string("\x45\xf0\x8f\x80\x7f\x45"), // UTF8-4 Invalid-2
		string("\x45\xf0\x8f\x7f\x80\x45"), // UTF8-4 Invalid-2
		string("\x45\xf0\x8f\x7f\x7f\x45"), // UTF8-4 Invalid-3
		string("\x45\xf0\x90\x7f\x45"),     // UTF8-4 Invalid-1-SubString
		string("\x45\xf0\x8f\x80\x45"),     // UTF8-4 Invalid-1-SubString
		string("\x45\xf0\x8f\x7f\x45"),     // UTF8-4 Invalid-2-SubString
		string("\x45\xf0\x8f\x45"),         // UTF8-4 Invalid-SubString

		string("\x45\xf1\x80\x80\x80\x45"), // UTF8-4 0xF1-0xF3 0x80-0xBF 0x80-0xBF 0x80-0xBF
		string("\x45\xf1\x80\x80\x45"),     // UTF8-4 SubString
		string("\x45\xf1\x80\x45"),         // UTF8-4 SubString
		string("\x45\xf1\x45"),             // UTF8-4 SubString
		string("\x45\xf1\x80\x80\x7f\x45"), // UTF8-4 Invalid-1
		string("\x45\xf1\x80\x7f\x80\x45"), // UTF8-4 Invalid-1
		string("\x45\xf1\x7f\x80\x80\x45"), // UTF8-4 Invalid-1
		string("\x45\xf1\x80\x7f\x7f\x45"), // UTF8-4 Invalid-2
		string("\x45\xf1\x7f\x80\x7f\x45"), // UTF8-4 Invalid-2
		string("\x45\xf1\x7f\x7f\x80\x45"), // UTF8-4 Invalid-2
		string("\x45\xf1\x7f\x7f\x7f\x45"), // UTF8-4 Invalid-3
		string("\x45\xf1\x80\x7f\x45"),     // UTF8-4 Invalid-1-SubString
		string("\x45\xf1\x7f\x80\x45"),     // UTF8-4 Invalid-1-SubString
		string("\x45\xf1\x7f\x7f\x45"),     // UTF8-4 Invalid-2-SubString
		string("\x45\xf1\x7f\x45"),         // UTF8-4 Invalid-SubString

		string("\x45\xf4\x80\x80\x80\x45"), // UTF8-4 0xF4 0x80-0x8F 0x80-0xBF 0x80-0xBF
		string("\x45\xf4\x80\x80\x45"),     // UTF8-4 SubString
		string("\x45\xf4\x80\x45"),         // UTF8-4 SubString
		string("\x45\xf4\x45"),             // UTF8-4 SubString
		string("\x45\xf4\x80\x80\x7f\x45"), // UTF8-4 Invalid-1
		string("\x45\xf4\x80\x7f\x80\x45"), // UTF8-4 Invalid-1
		string("\x45\xf4\x7f\x80\x80\x45"), // UTF8-4 Invalid-1
		string("\x45\xf4\x80\x7f\x7f\x45"), // UTF8-4 Invalid-2
		string("\x45\xf4\x7f\x80\x7f\x45"), // UTF8-4 Invalid-2
		string("\x45\xf4\x7f\x7f\x80\x45"), // UTF8-4 Invalid-2
		string("\x45\xf4\x7f\x7f\x7f\x45"), // UTF8-4 Invalid-3
		string("\x45\xf4\x80\x7f\x45"),     // UTF8-4 Invalid-1-SubString
		string("\x45\xf4\x7f\x80\x45"),     // UTF8-4 Invalid-1-SubString
		string("\x45\xf4\x7f\x7f\x45"),     // UTF8-4 Invalid-2-SubString
		string("\x45\xf4\x7f\x45"),         // UTF8-4 Invalid-SubString
	}

	for _, data := range datas {
		got := FormatStringToUtf8(data)
		if !utf8.ValidString(got) {
			t.Errorf("Fail to format to utf8: %s", data)
		}
	}
}
