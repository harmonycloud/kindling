// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package tools

const (
	AsciiLow     = byte(0x20)
	AsciiHigh    = byte(0x7e)
	AsciiReplace = byte(0x2e) // .
)

/*
 * Get the ascii readable string, replace other value to '.', like wireshark.
 */
func GetAsciiString(data []byte) string {
	length := len(data)
	if length == 0 {
		return ""
	}

	newData := make([]byte, length)
	for i := 0; i < length; i++ {
		if data[i] > AsciiHigh || data[i] < AsciiLow {
			newData[i] = AsciiReplace
		} else {
			newData[i] = data[i]
		}
	}
	return string(newData)
}
