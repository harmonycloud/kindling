// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package rocketmq

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
)

func TestParseRocketmqJsonAndRocketMQ(t *testing.T) {
	type args struct {
		message *protocol.PayloadMessage
	}
	reqs := [][]string{
		//JSON - short
		{"0000008a", "00000086|{\"code\":105,\"extFields\":{\"topic\":\"TopicTest\"},\"flag\":0,\"language\":\"JAVA\",\"opaque\":2034,\"serializeTypeCurrentRPC\":\"JSON\",\"version\":412}"},
		{"0000016e", "00000062|{\"code\":0,\"flag\":1,\"language\":\"JAVA\",\"opaque\":2034,\"serializeTypeCurrentRPC\":\"JSON\",\"version\":412}"},
		//JSON - long
		{"00000179", "00000164|{\"code\":310,\"extFields\":{\"a\":\"please_rename_unique_group_name\",\"b\":\"TopicTest\",\"c\":\"TBW102\",\"d\":\"4\",\"e\":\"1\",\"f\":\"0\",\"g\":\"1661601694866\",\"h\":\"0\",\"i\":\"UNIQ_KEY\\u00017F00000100D7135FBAA48879F8920026\\u0002WAIT\\u0001true\\u0002TAGS\\u0001TagA\",\"j\":\"0\",\"k\":\"false\",\"m\":\"false\"},\"flag\":0,\"language\":\"JAVA\",\"opaque\":84,\"serializeTypeCurrentRPC\":\"JSON\",\"version\":401}"},
		{"000000ec", "000000e8|{\"code\":0,\"extFields\":{\"queueId\":\"1\",\"TRACE_ON\":\"true\",\"MSG_REGION\":\"DefaultRegion\",\"msgId\":\"0AE95A5D00002A9F0000000000001C50\",\"queueOffset\":\"9\"},\"flag\":1,\"language\":\"JAVA\",\"opaque\":84,\"serializeTypeCurrentRPC\":\"JSON\",\"version\":401}"},
		//ROCKETMQ - short
		{"0000002d", "01000029", "006900019c", "0000000200000000", "0000000000000014", "0005|topic", "00000009|TopicTest"},
		{"00000122", "01000015", "000000019c", "0000000200000001", "0000000000000000", "7b2262726f6b65724461746173223a5b7b2262726f6b65724164647273223a7b2230223a223139322e3136382e36342e313a3130393131227d2c2262726f6b65724e616d65223a2262726f6b65722d61222c22636c7573746572223a2244656661756c74436c7573746572222c22656e61626c65416374696e674d6173746572223a66616c73657d5d2c2266696c7465725365727665725461626c65223a7b7d2c2271756575654461746173223a5b7b2262726f6b65724e616d65223a2262726f6b65722d61222c227065726d223a362c227265616451756575654e756d73223a342c22746f706963537973466c6167223a302c22777269746551756575654e756d73223a347d5d7d"},
		//ROCKETMQ - long
		{"0000011e", "01000108", "013600019c", "0000079c00000000", "00000000000000f3", "0001|a", "0000001f|please_rename_unique_group_name", "0001|b", "00000009|TopicTest", "0001|c", "00000006544257313032", "0001|d", "0000000134", "0001|e", "0000000130", "0001|f", "00000001300001670000000d31363633353837393039303934000168000000013000016900000055554e49515f4b4559014644453045363636363242463244443030303836383745413030373331374445304446353138423441414332363045463831453630334342025741495401747275650254414753015461674100016a000000013000016b0000000566616c736500016d0000000566616c736548656c6c6f20526f636b65744d5120393731"},
		{"000000e1", "010000dd", "000000019c", "0000079c00000001", "00000000000000c8", "00056d736749640000002043304138343030313030303032413946303030303030303030303139454544390007717565756549640000000130000b71756575654f66667365740000000431373432000d7472616e73616374696f6e4964000000384644453045363636363242463244443030303836383745413030373331374445304446353138423441414332363045463831453630334342000854524143455f4f4e0000000474727565000a4d53475f524547494f4e0000000d44656661756c74526567696f6e"},
	}
	datas := make([][]byte, len(reqs))
	for i, req := range reqs {
		datas[i] = getData(req)
	}
	tests := []struct {
		name       string
		args       args
		want       string
		wantHeader *rocketmqHeader
	}{
		//JSON
		{name: "json_request_short", args: args{message: &protocol.PayloadMessage{
			Data: datas[0],
		}}, want: "{\"code\":105,\"extFields\":{\"topic\":\"TopicTest\"},\"flag\":0,\"language\":\"JAVA\",\"opaque\":2034,\"serializeTypeCurrentRPC\":\"JSON\",\"version\":412}"},
		{name: "json_response_short", args: args{message: &protocol.PayloadMessage{
			Data: datas[1],
		}}, want: "{\"code\":0,\"flag\":1,\"language\":\"JAVA\",\"opaque\":2034,\"serializeTypeCurrentRPC\":\"JSON\",\"version\":412}"},
		{name: "json_request_long", args: args{message: &protocol.PayloadMessage{
			Data: datas[2],
		}}, want: "{\"code\":310,\"extFields\":{\"a\":\"please_rename_unique_group_name\",\"b\":\"TopicTest\",\"c\":\"TBW102\",\"d\":\"4\",\"e\":\"1\",\"f\":\"0\",\"g\":\"1661601694866\",\"h\":\"0\",\"i\":\"UNIQ_KEY\\u00017F00000100D7135FBAA48879F8920026\\u0002WAIT\\u0001true\\u0002TAGS\\u0001TagA\",\"j\":\"0\",\"k\":\"false\",\"m\":\"false\"},\"flag\":0,\"language\":\"JAVA\",\"opaque\":84,\"serializeTypeCurrentRPC\":\"JSON\",\"version\":401}"},
		{name: "json_response_long", args: args{message: &protocol.PayloadMessage{
			Data: datas[3],
		}}, want: "{\"code\":0,\"extFields\":{\"queueId\":\"1\",\"TRACE_ON\":\"true\",\"MSG_REGION\":\"DefaultRegion\",\"msgId\":\"0AE95A5D00002A9F0000000000001C50\",\"queueOffset\":\"9\"},\"flag\":1,\"language\":\"JAVA\",\"opaque\":84,\"serializeTypeCurrentRPC\":\"JSON\",\"version\":401}"},
		//ROCKETMQ
		{name: "rocketmq_request_short", args: args{message: &protocol.PayloadMessage{
			Data: datas[4],
		}}, wantHeader: &rocketmqHeader{
			Code:         105,
			Opaque:       2,
			Flag:         0,
			LanguageCode: 0,
		},
		},
		{name: "rocketmq_response_short", args: args{message: &protocol.PayloadMessage{
			Data: datas[5],
		}}, wantHeader: &rocketmqHeader{
			Code:         0,
			Opaque:       2,
			Flag:         1,
			LanguageCode: 0,
		},
		},
		{name: "rocketmq_request_long", args: args{message: &protocol.PayloadMessage{
			Data: datas[6],
		}}, wantHeader: &rocketmqHeader{
			Code:         310,
			Opaque:       1948,
			Flag:         0,
			LanguageCode: 0,
		},
		},
		{name: "rocketmq_response_long", args: args{message: &protocol.PayloadMessage{
			Data: datas[7],
		}}, wantHeader: &rocketmqHeader{
			Code:         0,
			Opaque:       1948,
			Flag:         1,
			LanguageCode: 0,
		},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				payLoadLength int32
				headerLength  int32
				serializeType uint8
			)
			tt.args.message.ReadInt32(0, &payLoadLength)

			header := &rocketmqHeader{ExtFields: map[string]string{}}
			if serializeType = tt.args.message.Data[4]; serializeType == 0 {
				tt.args.message.ReadInt32(4, &headerLength)
				_, headerBytes, err := tt.args.message.ReadBytes(8, int(headerLength))
				if err != nil {
					panic(err)
				}
				if err := json.Unmarshal(headerBytes, header); err != nil {
					panic(err)
				}
				fmt.Println(string(headerBytes))
				if string(headerBytes) != tt.want {
					t.Errorf("ReadBytes() = %v, want %v", string(headerBytes), tt.want)
				}
			} else if serializeType == 1 {
				parseHeader(tt.args.message, header)
				if header.Code != tt.wantHeader.Code {
					t.Errorf("header.code = %v, want %v", header.Code, tt.wantHeader.Code)
				}
				if header.LanguageCode != tt.wantHeader.LanguageCode {
					t.Errorf("header.languagecode = %v, want %v", header.LanguageCode, tt.wantHeader.LanguageCode)
				}
				if header.Opaque != tt.wantHeader.Opaque {
					t.Errorf("header.opaque = %v, want %v", header.Opaque, tt.wantHeader.Opaque)
				}
			}

			fmt.Printf("header code is %v \n", header.Code)
			fmt.Printf("header language is %v, languagecode is %v \n", rocketmqLanguageCode[header.LanguageCode], header.LanguageCode)
			fmt.Printf("header opaque is %v \n", header.Opaque)
			fmt.Printf("header flag is %v \n", header.Flag)
			fmt.Printf("header extFields is %v \n", header.ExtFields)
			//topicName maybe be stored in key `topic` or `b`
			if header.ExtFields["topic"] != "" {
				fmt.Printf("topic is \"%v\" \n", header.ExtFields["topic"])
			} else if header.ExtFields["b"] != "" {
				fmt.Printf("topic is \"%v\" \n", header.ExtFields["b"])
			} else if header.Flag == 0 {
				fmt.Printf("requestMsg is %v \n", requestMsgMap[header.Code])
			}
		})
	}

}

func getData(datas []string) []byte {
	dataBytes := make([]byte, 0)

	for _, data := range datas {
		if len(data) > 0 {
			dataSplit := getSplit(data)
			if dataSplit > 0 {
				sizeArray, _ := hex.DecodeString(data[0:dataSplit])
				dataBytes = append(dataBytes, sizeArray...)

				dataLen := len(data)
				for i := dataSplit + 1; i < dataLen; i++ {
					dataBytes = append(dataBytes, data[i])
				}
			} else if isHex(data[0]) {
				hexArray, _ := hex.DecodeString(data)
				dataBytes = append(dataBytes, hexArray...)
			} else {
				byteArray := []byte(data)
				dataBytes = append(dataBytes, byteArray...)
			}
		}
	}
	return dataBytes
}

func isHex(b byte) bool {
	if b >= '0' && b <= '9' {
		return true
	}
	if b >= 'a' && b <= 'z' {
		return true
	}
	return false
}

func getSplit(data string) int {
	if len(data) >= 3 && data[2] == '|' {
		return 2
	}
	if len(data) >= 5 && data[4] == '|' {
		return 4
	}
	if len(data) >= 9 && data[8] == '|' {
		return 8
	}
	return 0
}
