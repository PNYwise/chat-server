package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/IBM/sarama"
	_ "github.com/mattn/go-sqlite3"

	"github.com/PNYwise/chat-server/internal"
	"github.com/PNYwise/chat-server/internal/configs"
	"github.com/PNYwise/chat-server/internal/handler"
	"github.com/PNYwise/chat-server/internal/repository"
	chat_server "github.com/PNYwise/chat-server/proto"
	"google.golang.org/grpc"
)

func main() {

	// init configs
	internalConfig := configs.New()

	borker := fmt.Sprintf("%s:%d", internalConfig.GetString("kafka.host"), internalConfig.GetInt("kafka.port"))

	var brokerList []string = []string{borker}
	// kafka producer
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalf("Error closing Kafka producer: %v", err)
		}
	}()

	// kafka consumer
	consumer, err := sarama.NewConsumer(brokerList, config)
	if err != nil {
		log.Fatalf("Error creating Kafka consumer: %v", err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Fatalf("Error closing Kafka consumer: %v", err)
		}
	}()

	// Kafka topic to consume messages from
	partition := int32(0)
	offset := int64(sarama.OffsetNewest)

	// Create a partition consumer for the given topic, partition, and offset
	partitionConsumer, err := consumer.ConsumePartition(handler.TOPIC, partition, offset)
	if err != nil {
		log.Fatalf("Error creating partition consumer: %v", err)
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Fatalf("Error closing partition consumer: %v", err)
		}
	}()

	ctx := context.Background()

	// Initialize the db
	db := configs.DbConn(ctx, internalConfig)
	defer db.Close(ctx)

	// CREATE TABLE users and message IF EXIST.
	_, err = db.Exec(ctx, internal.CreateTableSQL)
	if err != nil {
		// Handle the error here, potentially including specific error checking
		log.Fatalf("Error creating tables: %v", err)
	} else {
		log.Println("Tables created successfully!")
	}

	// init repository
	userRepo := repository.NewUserRepository(db, ctx)
	messageRepo := repository.NewMessageRepository(db, ctx)

	// Inisialisasi server gRPC
	server := grpc.NewServer()

	chatServer := handler.NewChatHandler(producer, partitionConsumer, userRepo, messageRepo)
	// Register server gRPC
	chat_server.RegisterBroadcastServer(server, chatServer)

	appPort := fmt.Sprintf(":%d", internalConfig.GetInt("app.port"))
	// Run
	listener, err := net.Listen("tcp", appPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	log.Printf("Server listening on port %s \n", appPort)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v \n", err)
	}
}
