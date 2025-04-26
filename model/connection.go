package model

import "github.com/gocql/gocql"

type Connection struct {
	Id  gocql.UUID
	Did string
}
