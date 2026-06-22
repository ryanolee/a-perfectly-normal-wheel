package db

import _ "embed"

//go:embed schema.sql
var SchemaSQL string

//go:embed seed.sql
var SeedSQL string
