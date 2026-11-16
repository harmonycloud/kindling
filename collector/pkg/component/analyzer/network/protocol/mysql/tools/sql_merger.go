// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"regexp"
	"strings"
)

var (
	regexMatcher           = regexp.MustCompile("^[A-Za-z_-]+$")
	SQL_MERGER   SqlMerger = NewSqlMerger()
)

type SqlMerger struct {
	factory []SqlParser
}

type SqlParser struct {
	regexType *regexp.Regexp
	regexKey  *regexp.Regexp
	sqlType   string
	sqlKey    string
}

func NewSqlMerger() SqlMerger {
	factory := make([]SqlParser, 9)
	factory[0] = newSqlParser("select", "from")
	factory[1] = newSqlParser("insert", "into")
	factory[2] = newSqlParser("update", "update")
	factory[3] = newSqlParser("delete", "from")
	factory[4] = newSqlParser("drop", "index|table|database")
	factory[5] = newSqlParser("create", "index|table|database")
	factory[6] = newSqlParser("alter", "table")
	factory[7] = newSqlParser("set", "set")
	factory[8] = newSqlParser("commit", "commit")

	return SqlMerger{
		factory: factory,
	}
}

func (merger SqlMerger) ParseStatement(statement string) string {
	for _, parser := range merger.factory {
		if parser.checkType(statement) {
			result := parser.parseStatement(statement)
			if len(result) > 0 {
				return result
			}
		}
	}

	// Return empty when no matching
	return ""
}

func newSqlParser(sqlType string, sqlKey string) SqlParser {
	patternType := `(?i)(^\s*)` + sqlType + `(.*)`
	patternKey := `(?i)(?m)` + "(" + sqlKey + ")" + `(\s+(\S*)\s*|\n)`

	return SqlParser{
		regexType: regexp.MustCompile(patternType),
		regexKey:  regexp.MustCompile(patternKey),
		sqlType:   sqlType,
		sqlKey:    sqlKey,
	}
}

func getRegexMatchStr(sql string, reg *regexp.Regexp) string {
	matches := reg.FindStringSubmatch(sql)
	if len(matches) < 4 {
		return "*"
	}

	//Sql Statement[select * from table] or Regex[(from)\s+(\S*)\s?]
	// 0: from table
	// 1: from
	// 2: table
	// 3: table
	table := matches[3]
	if len(table) == 0 {
		return "*"
	}
	return table
}

func isSqlAlphabet(s string) bool {
	return regexMatcher.MatchString(s)
}

func (parser SqlParser) parseStatement(statement string) string {
	if !parser.checkType(statement) {
		return ""
	}

	value := getRegexMatchStr(statement, parser.regexKey)

	if !isSqlAlphabet(value) {
		// Converge the operated object in case of high cardinality
		value = "*"
	} else {
		value = value + " *"
	}
	return parser.sqlType + " " + value
}

func (parser SqlParser) checkType(statement string) bool {
	matches := parser.regexType.FindStringSubmatch(statement)
	if len(matches) == 0 {
		return false
	}

	return strings.Contains(strings.ToLower(matches[0]), parser.sqlType)
}
