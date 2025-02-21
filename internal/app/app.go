package app

import (
	grpcapp "github.com/azonnix/sso/internal/app/grpc"
	"github.com/azonnix/sso/internal/services/auth"
	sqlite "github.com/azonnix/sso/internal/storage/sqllite"
	"log/slog"
	"time"
)

type App struct {
	GRPCApp *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCApp: grpcApp,
	}
}
