package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	chat_server "github.com/PNYwise/chat-server/proto"
	"google.golang.org/grpc"
)

type ChatServer struct {
	clients      map[string]chat_server.Broadcast_CreateStreamServer
	clientsLock  sync.Mutex
	messageQueue chan *chat_server.Message
	chat_server.UnimplementedBroadcastServer
}

func (c *ChatServer) CreateStream(connect *chat_server.Connect, stream chat_server.Broadcast_CreateStreamServer) error {
	clientID := connect.GetId()

	c.clientsLock.Lock()
	c.clients[clientID] = stream
	c.clientsLock.Unlock()

	// TODO: check client if exist, if no insert new client

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

	// TODO CREATE TABLE users and message IF EXIST

	// Inisialisasi server gRPC
	server := grpc.NewServer()

	// Initialization ChatServer
	chatServer := &ChatServer{
		clients:      make(map[string]chat_server.Broadcast_CreateStreamServer),
		messageQueue: make(chan *chat_server.Message),
	}

	// Resgister server gRPC
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
