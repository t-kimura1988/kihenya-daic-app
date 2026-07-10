package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/sirupsen/logrus"
	"github.com/t-kimura1988/kihenya-daic-app/internal/db/sqlc"
)

var (
	db      *pgxpool.Pool
	queries *sqlc.Queries
	logger  *logrus.Logger
)

func init() {
	logger = logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
	logger.SetLevel(logrus.InfoLevel)
}

func buildConnStr() (string, error) {
	rdbUser := os.Getenv("K_DAIC_RDB_USER_NAME")
	rdbPassword := os.Getenv("K_DAIC_RDB_PASSWORD")
	rdbHost := os.Getenv("K_DAIC_RDB_HOST")
	rdbPort := os.Getenv("K_DAIC_RDB_PORT")
	rdbName := os.Getenv("K_DAIC_RDB_NAME")

	if rdbUser == "" || rdbPassword == "" {
		return "", fmt.Errorf("database credentials not loaded from environment variables")
	}

	if rdbHost == "" {
		rdbHost = "localhost"
	}
	if rdbPort == "" {
		rdbPort = "5432"
	}
	if rdbName == "" {
		rdbName = "k_daic_db"
	}

	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		rdbUser, rdbPassword, rdbHost, rdbPort, rdbName,
	), nil
}

func InitDB() error {
	ctx := context.Background()

	connStr, err := buildConnStr()
	if err != nil {
		return err
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("parse connection string: %w", err)
	}

	config.MaxConns = 5
	config.MinConns = 1
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	if config.ConnConfig.RuntimeParams == nil {
		config.ConnConfig.RuntimeParams = make(map[string]string)
	}
	config.ConnConfig.RuntimeParams["timezone"] = "Asia/Tokyo"

	config.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   tracelog.LoggerFunc(logTraceSQL),
		LogLevel: tracelog.LogLevelInfo,
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	db = pool
	queries = sqlc.New(db)
	logrus.Info("Database connected successfully")
	return nil
}

func CloseDB() {
	if db != nil {
		db.Close()
	}
}

func GetPgxConn() *pgxpool.Pool {
	return db
}

func GetDB() *sql.DB {
	return nil
}

func GetQueries() *sqlc.Queries {
	return queries
}

func logTraceSQL(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	if msg == "Query" {
		if query, ok := data["sql"].(string); ok {
			logger.WithFields(logrus.Fields{
				"type":  "SQL",
				"query": query,
				"args":  data["args"],
			}).Info("SQL Query Executed")
		}
	}
}
