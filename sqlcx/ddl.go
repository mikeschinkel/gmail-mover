package sqlcx

import (
	_ "embed"
)

//go:embed schema.sql
var ddl string

func DDL() string {
	return ddl
}
