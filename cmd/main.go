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
		msg := <-c.messageQueue
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
				userFrom := &internal.User{
					Id: uint(from),
				}
				userTo := &internal.User{
					Id: uint(to),
				}
				message := &internal.Message{Form: userFrom, To: userTo, Content: cpMsg.GetContent()}
				if err := c.messageRepo.Create(message); err != nil {
					log.Printf("Error sending queued message to client %s: %v", cpMsg.GetTo(), err)
				}
			}(msg)
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
	messageRepo := internal.NewMessageRepository(db)

	// Inisialisasi server gRPC
	server := grpc.NewServer()

	// Initialization ChatServer
	chatServer := &ChatServer{
		clients:      make(map[string]chat_server.Broadcast_CreateStreamServer),
		messageQueue: make(chan *chat_server.Message),
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
