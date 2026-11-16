// Copyright (C) 2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package model

import "github.com/Kindling-project/kindling/collector/pkg/model/constnames"

// ConvertSendmmsg converts the single sendmmsg event to multiple ones by parsing
// the data field. The original data is composed of the loop of "size(uint32)+data".
func ConvertSendmmsg(evt *KindlingEvent) []*KindlingEvent {
	ret := make([]*KindlingEvent, 0)
	if evt.Name != constnames.SendMMsgEvent {
		return ret
	}

	data := evt.GetData()
	bytesSlice := splitDataBytes(data)
	for _, bytes := range bytesSlice {
		evtTemplate := new(KindlingEvent)
		// Clone the original event
		*evtTemplate = *evt
		lenBytes := make([]byte, 8)
		byteOrder.PutUint64(lenBytes, uint64(len(bytes)))
		evtTemplate.SetUserAttribute("res", lenBytes)
		evtTemplate.SetUserAttribute("data", bytes)
		ret = append(ret, evtTemplate)
	}
	return ret
}

func splitDataBytes(data []byte) [][]byte {
	ret := make([][]byte, 0)
	dataLength := len(data)
	for startIndex := 0; startIndex < dataLength; {
		sizeEndIndex := startIndex + 4
		if sizeEndIndex >= dataLength {
			break
		}
		sizeBytes := data[startIndex:sizeEndIndex]
		size := byteOrder.Uint32(sizeBytes)
		fragmentEnd := sizeEndIndex + int(size)
		var childData []byte
		if fragmentEnd < dataLength {
			// Copy the array to avoid the slice being changed unexpectedly later
			childData = make([]byte, fragmentEnd-sizeEndIndex)
			copy(childData, data[sizeEndIndex:fragmentEnd])
		} else {
			// Copy the array to avoid the slice being changed unexpectedly later
			childData = make([]byte, dataLength-sizeEndIndex)
			copy(childData, data[sizeEndIndex:])
		}
		ret = append(ret, childData)
		startIndex = fragmentEnd
	}
	return ret
}
