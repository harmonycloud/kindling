// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package mysql

import (
	"encoding/binary"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

/*
int<3>	payload_length
int<1>	sequence_id
payload
*/
func fastfailMysqlResponse() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) < 6
	}
}

func parseMysqlResponse() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		return true, false
	}
}

/*
===== PayLoad =====
int<1>	header(0xff)
int<2>	error_code
if capabilities & CLIENT_PROTOCOL_41 {
	string[1]	sql state marker (#)
	string[5]	sql_state
}
string<EOF>	error_message
*/
func fastfailMysqlErr() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.Data[4] != 0xff
	}
}

func parseMysqlErr() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		errorCode := binary.LittleEndian.Uint16(message.Data[5:7])

		var errorMessage string
		if len(message.Data) > 14 && message.Data[8] == '#' {
			errorMessage = string(message.Data[8:13]) + ":" + string(message.Data[13:])
		} else {
			errorMessage = string(message.Data[8:])
		}

		message.AddIntAttribute(constlabels.SqlErrCode, int64(errorCode))
		message.AddUtf8StringAttribute(constlabels.SqlErrMsg, errorMessage)
		if errorCode != 0 {
			message.AddBoolAttribute(constlabels.IsError, true)
			message.AddIntAttribute(constlabels.ErrorType, int64(constlabels.ProtocolError))
		}
		return true, true
	}
}

/*
===== PayLoad =====
int<1>	header(0x00 or 0xFE)
int<lenenc>	affected_rows
int<lenenc>	last_insert_id
if capabilities & CLIENT_PROTOCOL_41 {
	int<2>	status_flags
	int<2>	warnings
} else if capabilities & CLIENT_TRANSACTIONS {
	int<2>	status_flags
}
if capabilities & CLIENT_SESSION_TRACK {
	string<lenenc>	info
	if status_flags & SERVER_SESSION_STATE_CHANGED {
		string<lenenc>	session state info
	}
} else {
	string<EOF>	info
}
*/
func fastfailMysqlOk() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.Data[4] != 0x00 && message.Data[4] != 0xfe
	}
}

func parseMysqlOk() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		return true, true
	}
}

/*
===== PayLoad =====
int<1>	header(0xFE)
if capabilities & CLIENT_PROTOCOL_41 {
	int<2>	warnings
	int<2>	status_flags
}
*/
func fastfailMysqlEof() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.Data[4] != 0xfe
	}
}

func parseMysqlEof() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		return true, true
	}
}

/*
if capabilities & CLIENT_OPTIONAL_RESULTSET_METADATA {
	int<1> metadata_follows
}
int<lenenc>	 column_count
if (not (capabilities & CLIENT_OPTIONAL_RESULTSET_METADATA)) or metadata_follows == RESULTSET_METADATA_FULL {
	column_count x Column Definition	field metadata
}
if (not capabilities & CLIENT_DEPRECATE_EOF) {
	EOF_Packet
}
One or more Text Resultset Row
if (error processing) {
	ERR_Packet
} else if capabilities & CLIENT_DEPRECATE_EOF {
	OK_Packet
} else {
	EOF_Packet
}
*/
func fastfailMysqlResultSet() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return !message.HasAttribute(constlabels.Sql)
	}
}

func parseMysqlResultSet() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		return true, true
	}
}
