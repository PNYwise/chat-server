package configs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
)

var dbPool *pgxpool.Pool

// initDatabasePool initializes the connection pool to the PostgreSQL database.
func initDatabasePool(ctx context.Context, config *viper.Viper) {
	username := config.GetString("database.username")
	password := config.GetString("database.password")
	host := config.GetString("database.host")
	port := config.GetInt("database.port")
	name := config.GetString("database.name")

	dbConfig := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		username,
		password,
		host,
		port,
		name,
	)

	poolConfig, err := pgxpool.ParseConfig(dbConfig)
	if err != nil {
		log.Fatalf("Unable to parse database config: %v", err)
	}

	// Best practice configurations
	poolConfig.MaxConns = 30                      // Maximum number of connections
	poolConfig.MinConns = 5                       // Minimum number of connections
	poolConfig.MaxConnLifetime = time.Hour        // Maximum connection lifetime
	poolConfig.MaxConnIdleTime = 15 * time.Minute // Maximum idle time for connections
	poolConfig.HealthCheckPeriod = time.Minute    // Health check interval

	// Initialize the connection pool
	dbPool, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}

	log.Println("Connected to Database")
}

// DbPool returns the database connection pool, initializing it if necessary.
func DbPool(ctx context.Context, config *viper.Viper) *pgxpool.Pool {
	if dbPool == nil {
		initDatabasePool(ctx, config)
	}
	return dbPool
}
