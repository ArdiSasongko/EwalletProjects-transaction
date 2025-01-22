package api

import (
	"context"

	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/config/db"
	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/config/logger"
	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/env"
	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/handler"

	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/storage/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func LoadConfig() (Config, error) {
	if err := godotenv.Load(); err != nil {
		return Config{}, nil
	}

	logrus := logger.NewLogger()

	cfg := Config{
		addrHTTP: env.GetEnvString("ADDR_HTTP", ":4000"),
		addrGRPC: env.GetEnvString("ADDR_GRPC", ":5000"),
		logger:   logrus,
		db: DBConfig{
			addr:         env.GetEnvString("DB_ADDR", ""),
			maxOpenConns: env.GetEnvInt("DB_MAX_CONNS", 5),
			maxIdleConns: env.GetEnvInt("DB_MAX_IDLE", 5),
			maxIdleTime:  env.GetEnvString("DB_MAX_TIME_IDLE", "10m"),
		},
		auth: AuthConfig{
			secret: env.GetEnvString("JWT_SECRET", ""),
			iss:    env.GetEnvString("JWT_ISS", ""),
			aud:    env.GetEnvString("JWT_AUD", ""),
		},
	}

	return cfg, nil
}

func ConnectDatabase(cfg DBConfig, logger *logrus.Logger) (*pgxpool.Pool, error) {
	conn, err := db.New(cfg.addr, cfg.maxOpenConns, cfg.maxIdleConns, cfg.maxIdleTime)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}

	logger.Info("success connected to database")
	return conn, nil
}

func SetupHTTPApplication() (*application, error) {
	cfg, err := LoadConfig()
	if err != nil {
		cfg.logger.Fatal(err.Error())
	}

	conn, err := ConnectDatabase(cfg.db, cfg.logger)
	if err != nil {
		cfg.logger.Fatalf("failed to connected database :%v", err)
	}

	//auth := auth.NewJwt(cfg.auth.secret, cfg.auth.aud, cfg.auth.iss)
	q := sqlc.New(conn)

	handler := handler.NewHandler(q, conn)

	return &application{
		config:  cfg,
		handler: handler,
	}, nil
}

func SetupGRPCApplication() (*application, error) {
	cfg, err := LoadConfig()
	if err != nil {
		cfg.logger.Fatal(err.Error())
	}
	return &application{
		config: cfg,
	}, nil
}
