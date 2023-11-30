package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/app"
	grpc "github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/app/grpc"
	rest "github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/app/http"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/config"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/store"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/models"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/pkg/shortener"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
	"github.com/sirupsen/logrus"
	grpc2 "google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"text/template"
)

// BuildInfo структура для хранения информации о сборке приложения
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

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

	if err := run(ctx); err != nil {
		panic("unexpected error: " + err.Error())
	}
}

func run(ctx context.Context) error {
	storage, err := newStore(ctx)
	if err != nil {
		return fmt.Errorf("cannot create storage: %w", err)
	}
	defer storage.Close()
	removeChan := make(chan models.BatchRemoveRequest)
	instance := app.NewInstance(config.BaseURL, storage, removeChan)
	restHandler := &rest.Handler{Instance: instance}

	grpcServer := grpc.NewShortenerServer(instance)
	s := grpc2.NewServer()
	shortener.RegisterShortenerServer(s, grpcServer)
	logrus.Printf("Starting gRPC server on port: %v", config.GrpcPort)
	lis, err := net.Listen("tcp", config.GrpcPort)
	if err != nil {
		logrus.Fatalf("grpc listen error: %v", err)
	}

	go func() {
		err = s.Serve(lis)
		if err != nil {
			logrus.Fatalf("grpc serve error: %v", err)
		}
	}()

	wg := sync.WaitGroup{}
	go func() {
		for removeRequest := range removeChan {
			wg.Add(1)
			go func(req models.BatchRemoveRequest) {
				defer wg.Done()
				err := storage.DeleteUsers(ctx, req.UID, req.Ids...)
				if err != nil {
					logrus.Errorf("Couldn't delete urls for user %s", req.UID.String())
				}
			}(removeRequest)
		}
	}()
	srv := &http.Server{
		Addr:    config.RunPort,
		Handler: newRouter(restHandler),
	}

	if config.UseTLS {
		err := config.MakeKeys(config.CertFile, config.KeyFile)
		if err != nil {
			return err
		}
		srv.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS13,
		}
		go func() {
			if err := srv.ListenAndServeTLS(config.CertFile, config.KeyFile); err != nil && err != http.ErrServerClosed {
				logrus.Fatalf("listen and serve: %v", err)
			}
		}()
	} else {
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logrus.Fatalf("listen and serve: %v", err)
			}
		}()
	}

	logrus.Infof("listening on %s", config.RunPort)

	<-ctx.Done()

	logrus.Info("shutting down server gracefully")
	idleConnectionsClosed := make(chan struct{})

	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	go func() {
		var wait sync.WaitGroup
		wait.Add(2)
		go func() {
			defer wait.Done()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				logrus.Errorf("shutdown: %v", err)
			}
			logrus.Println("HTTP server stopped.")
		}()
		go func() {
			defer wait.Done()
			s.GracefulStop()
			logrus.Println("GRPC server stopped.")
		}()
		wait.Wait()
		close(idleConnectionsClosed)
	}()

	select {
	case <-shutdownCtx.Done():
		return fmt.Errorf("server shutdown: %w", shutdownCtx.Err())
	case <-idleConnectionsClosed:
		logrus.Println("servers stopped.")
	}

	wg.Wait()
	logrus.Println("All processes done.")
	return nil
}

func newStore(ctx context.Context) (storage store.AuthStore, err error) {
	if config.DatabaseDSN != "" {
		logrus.Debug("Create DB storage")
		rdb, errs := newRDBStore(ctx, config.DatabaseDSN)
		if errs != nil {
			return nil, fmt.Errorf("cannot create RDB store: %w", err)
		}
		if err = rdb.Bootstrap(ctx); err != nil {
			return nil, fmt.Errorf("cannot bootstrap RDB store: %w", err)
		}
		return rdb, nil
	}
	if config.PersistFile != "" {
		logrus.Debug("Create file storage")
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
