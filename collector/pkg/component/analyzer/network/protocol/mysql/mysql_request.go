// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package mysql

import (
	"strings"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol/mysql/tools"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

/*
int<3>	payload_length
int<1>	sequence_id
payload
*/
func fastfailMysqlRequest() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) < 5
	}
}

func parseMysqlRequest() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		return true, false
	}
}

/*
===== PayLoad =====
1              COM_STMT_PREPARE<0x16>
string[EOF]    the query to prepare
*/
func fastfailMysqlPrepare() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.Data[4] != 22
	}
}

func parseMysqlPrepare() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		sql := string(message.Data[5:])
		if !isSql(sql) {
			return false, true
		}
		message.AddUtf8StringAttribute(constlabels.Sql, sql)
		message.AddUtf8StringAttribute(constlabels.ContentKey, tools.SQL_MERGER.ParseStatement(sql))
		return true, true
	}
}

/*
===== PayLoad =====
1              COM_QUERY<03>
CLIENT_QUERY_ATTRIBUTES

	1            Number of parameters
	1            Number of parameter sets. Currently always 1
	...

string[EOF]    the query the server shall execute
*/
func fastfailMysqlQuery() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.Data[4] != 3
	}
}

func parseMysqlQuery() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		sql := string(message.Data[5:])
		if len(sql) > 2 && sql[0] == 0x00 && sql[1] == 0x01 {
			// Only Fix Zero params Case.
			// TODO Fix One more params case.
			sql = sql[2:]
		}
		if !isSql(sql) {
			return false, true
		}

		message.AddUtf8StringAttribute(constlabels.Sql, sql)
		message.AddUtf8StringAttribute(constlabels.ContentKey, tools.SQL_MERGER.ParseStatement(sql))
		return true, true
	}
}

/*
===== PayLoad =====
1              COM_QUIT<01>
*/
func fastfailMysqlQuit() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.Data[4] != 1
	}
}

func parseMysqlQuit() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		message.AddBoolAttribute(constlabels.Oneway, true)
		return true, true
	}
}

var sqlPrefixs = []string{
	"select",
	"insert",
	"update",
	"delete",
	"drop",
	"create",
	"alter",
	"set",
	"commit",
}

func isSql(sql string) bool {
	lowerSql := strings.ToLower(sql)

	for _, prefix := range sqlPrefixs {
		if strings.HasPrefix(lowerSql, prefix) {
			return true
		}
	}
	return false
}
