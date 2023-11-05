package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/app"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/config"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/store"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
	"net/http"
	"os"
	"text/template"
	"time"
)

type BuildInfo struct {
	BuildVersion string
	BuildDate    string
	BuildCommit  string
}

var buildVersion string
var buildDate string
var buildCommit string

const buildInfo = `Build version: {{if .BuildVersion}}{{.BuildVersion}}{{else}}N/A{{end}}
Build date: {{if .BuildDate}}{{.BuildDate}}{{else}}N/A{{end}}
Build commit: {{if .BuildCommit}}{{.BuildCommit}}{{else}}N/A{{end}}
`

func main() {
	bi := BuildInfo{
		BuildVersion: buildVersion,
		BuildDate:    buildDate,
		BuildCommit:  buildCommit,
	}

	t := template.Must(template.New("list").Parse(buildInfo))
	err := t.Execute(os.Stdout, bi)
	if err != nil {
		panic(err)
	}

	config.Parse()

	if err := run(); err != nil {
		panic("unexpected error: " + err.Error())
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	storage, err := newStore(ctx)
	if err != nil {
		return fmt.Errorf("cannot create storage: %w", err)
	}
	defer storage.Close()

	instance := app.NewInstance(config.BaseURL, storage)

	return http.ListenAndServe(config.RunPort, newRouter(instance))
}

func newStore(ctx context.Context) (storage store.AuthStore, err error) {
	if config.DatabaseDSN != "" {
		rdb, err := newRDBStore(ctx, config.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("cannot create RDB store: %w", err)
		}
		if err := rdb.Bootstrap(ctx); err != nil {
			return nil, fmt.Errorf("cannot bootstrap RDB store: %w", err)
		}
		return rdb, nil
	}
	if config.PersistFile != "" {
		storage, err = store.NewFileStore(config.PersistFile)
		if err != nil {
			return nil, fmt.Errorf("cannot create file store: %w", err)
		}
		return
	}
	return store.NewInMemory(), nil
}

func newRDBStore(ctx context.Context, dsn string) (*store.RDB, error) {
	// disable prepared statements
	driverConfig := stdlib.DriverConfig{
		ConnConfig: pgx.ConnConfig{
			PreferSimpleProtocol: true,
		},
	}
	stdlib.RegisterDriverConfig(&driverConfig)

	conn, err := sql.Open("pgx", driverConfig.ConnectionString(dsn))
	if err != nil {
		return nil, fmt.Errorf("cannot create connection pool: %w", err)
	}

	if err = conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("cannot perform initial ping: %w", err)
	}

	return store.NewRDB(conn), nil
}
