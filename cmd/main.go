package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5"
	_ "github.com/mattn/go-sqlite3"

	"github.com/PNYwise/chat-server/internal"
	chat_server "github.com/PNYwise/chat-server/proto"
	"google.golang.org/grpc"
)

const topic = "chat"

type ChatServer struct {
	clients      map[string]chat_server.Broadcast_CreateStreamServer
	clientsLock  sync.Mutex
	messageQueue chan *chat_server.Message
	producer     sarama.SyncProducer
	consumer     sarama.PartitionConsumer
	userRepo     internal.IUserRepository
	messageRepo  internal.IMessageRepository
	chat_server.UnimplementedBroadcastServer
}

func (c *ChatServer) CreateStream(connect *chat_server.Connect, stream chat_server.Broadcast_CreateStreamServer) error {
	stringClientID := connect.GetId()

	clientID, err := strconv.Atoi(stringClientID)
	if err != nil {
		return err
	}
	c.clientsLock.Lock()
	//check client if exist, if no insert new client
	if ok := c.userRepo.Exist(clientID); !ok {
		user := &internal.User{
			Name: connect.Name,
		}
		if err := c.userRepo.Create(user); err != nil {
			return err
		}
		stringClientID = strconv.Itoa(int(user.Id))
	}

	// add client connection in memory
	c.clients[stringClientID] = stream

	// check existing message when client offline
	messages, err := c.messageRepo.ReadByUserId(uint(clientID))
	if err != nil {
		return err
	}
	// send existing message
	if len(*messages) > 0 {
		var messageId []uint
		for _, v := range *messages {
			messageId = append(messageId, v.Id)

			msg := &chat_server.Message{
				From:    strconv.Itoa(int(v.Form.Id)),
				To:      strconv.Itoa(int(v.To.Id)),
				Content: v.Content,
			}
			if err := stream.Send(msg); err != nil {
				if err := c.messageRepo.Create(&v); err != nil {
					log.Printf("Error sending queued message to client %d: %v", v.To.Id, err)
				}
			}
		}
		if err := c.messageRepo.Delete(messageId); err != nil {
			return err
		}
	}
	c.clientsLock.Unlock()

	for {
		byteMessage := <-c.consumer.Messages()

		msg := new(chat_server.Message)
		if err := json.Unmarshal(byteMessage.Value, msg); err != nil {
			log.Printf("Error sending queued message to client %s: %v", msg.GetTo(), err)
		}

		if err := c.sendMessageToClient(msg.GetTo(), msg); err != nil {
			//save unsed messsage
			go func(cpMsg *chat_server.Message) {
				from, err := strconv.Atoi(cpMsg.GetFrom())
				if err != nil {
					log.Printf("Error sending queued message to client %s: %v", cpMsg.GetTo(), err)
				}
				to, err := strconv.Atoi(cpMsg.GetTo())
				if err != nil {
					log.Printf("Error sending queued message to client %s: %v", cpMsg.GetTo(), err)
				}

				userFrom := &internal.User{Id: uint(from)}
				userTo := &internal.User{Id: uint(to)}

				message := &internal.Message{Form: userFrom, To: userTo, Content: cpMsg.GetContent()}
				if err := c.messageRepo.Create(message); err != nil {
					log.Printf("Error sending queued message to client %s: %v", cpMsg.GetTo(), err)
				}
			}(msg)
		}
	}
}

func (c *ChatServer) sendMessageToClient(clientID string, msg *chat_server.Message) error {
	c.clientsLock.Lock()
	defer c.clientsLock.Unlock()
	clientStream, ok := c.clients[clientID]
	if !ok {
		return nil
	}

	if err := clientStream.Send(msg); err != nil {
		return err
	}

	return nil
}

func (c *ChatServer) BroadcastMessage(ctx context.Context, message *chat_server.Message) (*chat_server.Close, error) {
	toId, err := strconv.Atoi(message.GetTo())
	if err != nil {
		return nil, err
	}
	exist := c.userRepo.Exist(toId)
	if !exist {
		return nil, errors.New("user not found")
	}
	go func() {
		if err := c.publishMessage(topic, message); err != nil {
			fmt.Printf("err send message :%v", err)
		}
	}()
	return &chat_server.Close{}, nil
}

func (c *ChatServer) publishMessage(topic string, message *chat_server.Message) error {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}
	_, _, err = c.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(jsonMessage),
	})
	return err
}

func main() {
	var brokerList []string = []string{"127.0.0.1:9092"}
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
	partitionConsumer, err := consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		log.Fatalf("Error creating partition consumer: %v", err)
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Fatalf("Error closing partition consumer: %v", err)
		}
	}()

	ctx := context.Background()
	// Initialization GB
	dbConfig := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		"user",
		"password",
		"localhost",
		54321,
		"post-chat",
	)
	connConfig, err := pgx.ParseConfig(dbConfig)
	if err != nil {
		log.Fatal(err)
	}

	db, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to Database")

	// CREATE TABLE users and message IF EXIST.
	_, err = db.Exec(ctx, internal.CreateTableSQL)
	if err != nil {
		// Handle the error here, potentially including specific error checking
		log.Fatalf("Error creating tables: %v", err)
	} else {
		log.Println("Tables created successfully!")
	}
	// init repository
	userRepo := internal.NewUserRepository(db, ctx)
	messageRepo := internal.NewMessageRepository(db,ctx)

	// Inisialisasi server gRPC
	server := grpc.NewServer()

	// Initialization ChatServer
	chatServer := &ChatServer{
		clients:      make(map[string]chat_server.Broadcast_CreateStreamServer),
		messageQueue: make(chan *chat_server.Message),
		producer:     producer,
		consumer:     partitionConsumer,
		userRepo:     userRepo,
		messageRepo:  messageRepo,
	}

	// Register server gRPC
	chat_server.RegisterBroadcastServer(server, chatServer)

	// Run
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	log.Println("Server listening on port :50051")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
