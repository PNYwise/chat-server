package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"strconv"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/PNYwise/chat-server/internal"
	chat_server "github.com/PNYwise/chat-server/proto"
	"google.golang.org/grpc"
)

type ChatServer struct {
	clients      map[string]chat_server.Broadcast_CreateStreamServer
	clientsLock  sync.Mutex
	messageQueue chan *chat_server.Message
	userRepo     internal.IUserRepository
	chat_server.UnimplementedBroadcastServer
}

func (c *ChatServer) CreateStream(connect *chat_server.Connect, stream chat_server.Broadcast_CreateStreamServer) error {
	stringClientID := connect.GetId()

	clientID, err := strconv.Atoi(stringClientID)
	if err != nil {
		return err
	}
	//check client if exist, if no insert new client
	if ok := c.userRepo.Exist(clientID); !ok {
		return err
	}
	c.clientsLock.Lock()
	c.clients[stringClientID] = stream
	c.clientsLock.Unlock()

	// TODO: check existing message when client offline

	// TODO: send existing message
	for {
		msg := <-c.messageQueue
		if err := c.sendMessageToClient(msg.GetTo(), msg); err != nil {
			log.Printf("Error sending queued message to client %s: %v", msg.GetTo(), err)
		}
	}
}

func (s *ChatServer) sendMessageToClient(clientID string, msg *chat_server.Message) error {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()
	clientStream, ok := s.clients[clientID]
	if !ok {
		return nil
	}

	if err := clientStream.Send(msg); err != nil {
		return err
	}

	return nil
}

func (c *ChatServer) BroadcastMessage(ctx context.Context, message *chat_server.Message) (*chat_server.Close, error) {
	// TODO: check receiver if exist

	go func() {
		c.messageQueue <- message
	}()
	return &chat_server.Close{}, nil
}

func main() {
	// Initialization GB
	db, err := sql.Open("sqlite3", "chat.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	// CREATE TABLE users and message IF EXIST.
	ctx := context.Background()
	_, err = db.ExecContext(ctx, internal.CreateTableSQL)
	if err != nil {
		// Handle the error here, potentially including specific error checking
		log.Fatalf("Error creating tables: %v", err)
	} else {
		log.Println("Tables created successfully!")
	}
	// init repository
	userRepo := internal.NewUserRepository(db)

	// Inisialisasi server gRPC
	server := grpc.NewServer()

	// Initialization ChatServer
	chatServer := &ChatServer{
		clients:      make(map[string]chat_server.Broadcast_CreateStreamServer),
		messageQueue: make(chan *chat_server.Message),
		userRepo:     userRepo,
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
