package postgreshelper

import (
	"context"
	"database/sql"
	"fmt"

	"brandlovrs.exporter/db/hotdata"
	"brandlovrs.exporter/internal/config"
	_ "github.com/lib/pq"
)

// Connection parameters
const (
	host     = "hotdata.catalystcore.io"
	port     = 5439
	dbname   = "db-hotdata-hmlg"
	user     = "postgres"
	password = "H0t!Dt4r#UYm7iEy27%hjrFboZ"
)

var (
	db      *sql.DB
	err     error
	ctx     context.Context
	Queries *hotdata.Queries
)

func init() {
	ctx = getContext()
	db, err = PostgressConnector(ctx)
	Queries = hotdata.New(db)
}

func getContext() context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return ctx
}

func AuroraConnector(ctx context.Context) (*sql.DB, error) {
	// Construct connection string
	connStr := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=require", host, port, dbname, user, password)

	println(connStr)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func PostgressConnector(ctx context.Context) (*sql.DB, error) {

	db, err := sql.Open("postgres", config.ReadParameter("POSTGRESQL_URI"))
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db, nil
}
