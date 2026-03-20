package main

import (
	"context"
	"database/sql"
)

type DatabaseContext struct {
	ctx context.Context
	db  *sql.DB
}

func main() {

	ctx := context.Background()

	if ctx == nil {
		panic("context is empty")
	}

}
