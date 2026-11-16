// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package mysql

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
)

/*
		    Request                                       Response
		/     |       \                                 /     |    \
	 prepare query   quit                              err   ok    eof
*/
func NewMysqlParser() *protocol.ProtocolParser {
	requestParser := protocol.CreatePkgParser(fastfailMysqlRequest(), parseMysqlRequest())
	requestParser.Add(fastfailMysqlPrepare(), parseMysqlPrepare())
	requestParser.Add(fastfailMysqlQuery(), parseMysqlQuery())
	requestParser.Add(fastfailMysqlQuit(), parseMysqlQuit())

	responseParser := protocol.CreatePkgParser(fastfailMysqlResponse(), parseMysqlResponse())
	responseParser.Add(fastfailMysqlErr(), parseMysqlErr())
	responseParser.Add(fastfailMysqlOk(), parseMysqlOk())
	responseParser.Add(fastfailMysqlEof(), parseMysqlEof())
	responseParser.Add(fastfailMysqlResultSet(), parseMysqlResultSet())

	return protocol.NewProtocolParser(protocol.MYSQL, requestParser, responseParser, nil)
}
