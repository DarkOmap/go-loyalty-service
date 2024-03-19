package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"

	"github.com/Tomap-Tomap/go-loyalty-service/iternal/handlers"
	"github.com/Tomap-Tomap/go-loyalty-service/iternal/logger"
	"github.com/Tomap-Tomap/go-loyalty-service/iternal/parameters"
	"github.com/Tomap-Tomap/go-loyalty-service/iternal/storage"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func main() {
	p := parameters.ParseFlags()

	if err := logger.Initialize("INFO", "stderr"); err != nil {
		panic(err)
	}

	logger.Log.Info("Create database storage")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	conn, err := pgx.Connect(ctx, p.DataBaseURI)

	if err != nil {
		logger.Log.Fatal("Connect to database", zap.Error(err))
	}
	defer conn.Close(ctx)

	storage, err := storage.NewStorage(conn)

	if err != nil {
		logger.Log.Fatal("Create storage", zap.Error(err))
	}

	logger.Log.Info("Create handlers")
	h := handlers.NewHandlers(*storage)
	logger.Log.Info("Create mux")
	mux := handlers.ServiceMux(h)

	if err := http.ListenAndServe(p.RunAddr, mux); err != nil {
		logger.Log.Fatal("Problem with working server", zap.Error(err))
	}
}
