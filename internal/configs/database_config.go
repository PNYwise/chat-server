package configs

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
)

var db *pgx.Conn

func getDatabaseConn(ctx context.Context, config *viper.Viper) *pgx.Conn {

	username := config.GetString("database.username")
	password := config.GetString("database.password")
	host := config.GetString("database.host")
	port := config.GetInt("database.port")
	name := config.GetString("database.name")

	dbConfig := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		username,
		password,
		host,
		port,
		name,
	)
	connConfig, err := pgx.ParseConfig(dbConfig)
	if err != nil {
		log.Fatal(err)
	}

	newDB, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		log.Fatal(err)
	}

	if err := newDB.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to Database")
	db = newDB
	return db
}
func DbConn(ctx context.Context, config *viper.Viper) *pgx.Conn {
	if db == nil {
		db = getDatabaseConn(ctx, config)
	}
	return db
}
