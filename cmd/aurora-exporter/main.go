// Command aurora-exporter exports query results from Redshift (or any
// supported SQL source) into PostgreSQL, Aurora, MySQL and other
// targets, on a schedule or on demand.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Jeffersonmf/aurora-exporter/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	code := cli.Main(ctx, os.Args[1:])
	stop()
	os.Exit(code)
}
