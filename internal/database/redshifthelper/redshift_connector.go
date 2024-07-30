package redshifthelper

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Connection parameters
const (
	host     = "datalake-brandlovrs.ct1jliawl8po.us-east-2.redshift.amazonaws.com"
	port     = 5439
	dbname   = "datahub"
	user     = "jefferson"
	password = "Admin*123"
)

func RedshiftConnector(ctx context.Context) (*sql.DB, error) {

	// Construct connection string
	connStr := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=require", host, port, dbname, user, password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db, nil
}
