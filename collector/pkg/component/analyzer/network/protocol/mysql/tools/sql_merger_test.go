// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"testing"
)

func TestSqlMerger_InsertSql(t *testing.T) {
	sqlMerger := NewSqlMerger()

	tests := []struct {
		operator string
		datas    map[string][]string
	}{
		{
			operator: "insert",
			datas: map[string][]string{
				"insert Websites *": {
					"INSERT INTO Websites (name, url, alexa, country)" +
						"VALUES ('baidu','https://www.baidu.com/','4','CN');",
				},
			},
		},
		{
			operator: "create",
			datas: map[string][]string{
				"create Persons *": {
					"CREATE table Persons\n" +
						"(\n" +
						"PersonID int,\n" +
						"LastName varchar(255),\n" +
						"FirstName varchar(255),\n" +
						"Address varchar(255),\n" +
						"City varchar(255)\n" +
						")",
				},
				"create *": {
					"CREATE DATABASE dbname;",
				},
				"create PIndex *": {
					"CREATE INDEX PIndex\nON Persons (LastName)",
				},
			},
		},
		{
			operator: "select",
			datas: map[string][]string{
				"select table *": {
					"select * from table",
					"select * from table ",
					"select * from table where id = 1",
				},
				"select person *": {
					"select name \n" +
						"from person \n" +
						"where countryid in ( select countryid \n" +
						"                     from country\n" +
						"                     where countryname = 'china');",
				},
				"select *": {
					"SELECT A.SERVICE\n" +
						"FROM (SELECT SERVICE_NAME AS SERVICE, COUNT(CODE) CODE_COUNT\n" +
						"   FROM log_detail\n" +
						"   GROUP BY SERVICE_NAME\n" +
						"   ORDER BY CODE DESC LIMIT 0,10) as A",
					"select a2333",
				},
			},
		},
		{
			operator: "delete",
			datas: map[string][]string{
				"delete Websites *": {
					"DELETE FROM Websites\nWHERE name='Facebook' AND country='USA';",
				},
			},
		},
		{
			operator: "commit",
			datas: map[string][]string{
				"commit *": {
					"commit",
				},
			},
		},
		{
			operator: "set",
			datas: map[string][]string{
				"set *": {
					"SET autocommit=0",
					"SET autocommit=1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.operator, func(t *testing.T) {
			for want, sqls := range tt.datas {
				for _, sql := range sqls {
					if got := sqlMerger.ParseStatement(sql); got != want {
						t.Errorf("ParseStatement() = %v, want %v", got, want)
					}
				}
			}
		})
	}
}
