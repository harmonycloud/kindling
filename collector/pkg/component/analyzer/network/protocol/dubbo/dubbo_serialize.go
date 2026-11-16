// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package dubbo

const (
	JsonNextLine = byte(0x0a)
	JsonQutoes   = byte(0x22)
	JsonColon    = byte(0x3a)

	SerialHessian2 = byte(0x02)
	SerialFastjson = byte(0x06)
)

type dubboSerializer interface {
	eatString(data []byte, offset int) int

	getStringValue(data []byte, offset int) (int, string)

	getStringValueByKey(data []byte, offset int, key string) string
}

var (
	serialHessian2  = &dubboHessian{}
	serialFastjson  = &dubboFastJson{}
	serialUnsupport = &dubboUnsupport{}
)

func GetSerializer(serialID byte) dubboSerializer {
	switch serialID {
	case SerialHessian2:
		return serialHessian2
	case SerialFastjson:
		return serialFastjson
	default:
		return serialUnsupport
	}
}

type dubboHessian struct{}

func (dh *dubboHessian) eatString(data []byte, offset int) int {
	dataLength := len(data)
	if offset >= dataLength {
		return dataLength
	}

	tag := data[offset]
	if tag >= 0x30 && tag <= 0x33 {
		if offset+1 == dataLength {
			return dataLength
		}
		// [x30-x34] <utf8-data>
		return offset + 2 + int(tag-0x30)<<8 + int(data[offset+1])
	} else {
		return offset + 1 + int(tag)
	}
}

func (dh *dubboHessian) getStringValue(data []byte, offset int) (int, string) {
	dataLength := len(data)
	if offset >= dataLength {
		return dataLength, ""
	}

	var stringValueLength int
	tag := data[offset]
	if tag >= 0x30 && tag <= 0x33 {
		if offset+1 == dataLength {
			return dataLength, ""
		}
		// [x30-x34] <utf8-data>
		stringValueLength = int(tag-0x30)<<8 + int(data[offset+1])
		offset += 2
	} else {
		stringValueLength = int(tag)
		offset += 1
	}

	if offset+stringValueLength >= len(data) {
		return dataLength, string(data[offset:])
	}
	return offset + stringValueLength, string(data[offset : offset+stringValueLength])
}

func (dh *dubboHessian) getStringValueByKey(data []byte, from int, key string) string {
	keyLength := len(key)
	dataLength := len(data)
	firstKeyword := byte(key[0])

	for i := from; i < dataLength; i++ {
		if data[i] == firstKeyword {
			matchKey := dh.getStrValue(data, dataLength, i, keyLength)
			if matchKey == key {
				_, value := dh.getStringValue(data, i+keyLength)
				return value
			}
		}
	}
	return ""
}

func (dh *dubboHessian) getStrValue(data []byte, dataLength int, index int, length int) string {
	if index >= dataLength {
		return ""
	}
	if index+length >= len(data) {
		// Not Enough String, Skip it.
		return ""
	}
	return string(data[index : index+length])
}

type dubboFastJson struct{}

func (json *dubboFastJson) eatString(data []byte, offset int) int {
	dataLength := len(data)
	if offset >= dataLength {
		return dataLength
	}

	/*
	    "xxx"\n
	    |    |
	   off   i
	*/
	for i := offset + 1; i < dataLength; i++ {
		if data[i] == JsonNextLine {
			return i + 1
		}
	}
	return dataLength
}

func (json *dubboFastJson) getStringValue(data []byte, offset int) (int, string) {
	dataLength := len(data)
	if offset >= dataLength {
		return dataLength, ""
	}

	/*
	    "xxx"\n
	    |    |
	   off   i
	*/
	for i := offset + 1; i < dataLength; i++ {
		if data[i] == JsonNextLine {
			return i + 1, string(data[offset+1 : i-1])
		}
	}
	return dataLength, ""
}

func (json *dubboFastJson) getStringValueByKey(data []byte, from int, key string) string {
	keyLength := len(key)
	dataLength := len(data)

	/*
	  "keyxxxxxxxx":"value"
	  |           |
	  quoteLeft   i
	*/
	quoteLeft := 0
	for i := from; i < dataLength; i++ {
		if data[i] == JsonQutoes {
			if quoteLeft == 0 {
				// Set Left Index
				quoteLeft = i
			} else if data[i+1] == JsonColon && data[i+2] == JsonQutoes {
				// "key":"value"
				if i-quoteLeft-1 == keyLength && string(data[quoteLeft+1:i]) == key {
					return json.getNextString(data, dataLength, i+2)
				}
				quoteLeft = 0
			} else {
				// Rest to zero
				quoteLeft = 0
			}
		}
	}
	return ""
}

func (json *dubboFastJson) getNextString(data []byte, dataLength int, offset int) string {
	if offset >= dataLength {
		return ""
	}

	/*
	    "xxx"
	    |   |
	   off  i
	*/
	for i := offset + 1; i < dataLength; i++ {
		if data[i] == JsonQutoes {
			return string(data[offset+1 : i])
		}
	}
	// Not Enough String, Skip it.
	return ""
}

type dubboUnsupport struct{}

func (unsupport *dubboUnsupport) eatString(data []byte, offset int) int {
	return 0
}

func (unsupport *dubboUnsupport) getStringValue(data []byte, offset int) (int, string) {
	return 0, ""
}

func (unsupport *dubboUnsupport) getStringValueByKey(data []byte, offset int, key string) string {
	return ""
}
