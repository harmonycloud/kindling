// Copyright (C) 2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cameraexporter

import (
	"strconv"
	"time"
)

func getDateString(timestamp int64) string {
	timeUnix := time.Unix(0, timestamp)
	year, month, day := timeUnix.Date()
	hour, minute, second := timeUnix.Clock()
	nano := timestamp % 1e9
	return dateToString(year) + dateToString(int(month)) + dateToString(day) +
		dateToString(hour) + dateToString(minute) + dateToString(second) +
		"." + dateToString(int(nano))
}

func dateToString(date int) string {
	if date >= 0 && date <= 9 {
		return "0" + strconv.FormatInt(int64(date), 10)
	} else {
		return strconv.FormatInt(int64(date), 10)
	}
}
