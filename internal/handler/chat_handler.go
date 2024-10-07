package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/PNYwise/chat-server/internal/domain"
	chat_server "github.com/PNYwise/chat-server/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ChatHandler struct {
	clients     map[string]chat_server.Broadcast_CreateStreamServer
	clientsLock sync.Mutex
	config      *viper.Viper
	producer    sarama.SyncProducer
	consumer    sarama.PartitionConsumer
	userRepo    domain.IUserRepository
	messageRepo domain.IMessageRepository
	chat_server.UnimplementedBroadcastServer
}

func NewChatHandler(config *viper.Viper, producer sarama.SyncProducer, consumer sarama.PartitionConsumer, userRepo domain.IUserRepository, messageRepo domain.IMessageRepository) *ChatHandler {
	return &ChatHandler{
		clients:     make(map[string]chat_server.Broadcast_CreateStreamServer),
		config:      config,
		producer:    producer,
		consumer:    consumer,
		userRepo:    userRepo,
		messageRepo: messageRepo,
	}
}

func (c *ChatHandler) CreateStream(_ *empty.Empty, stream chat_server.Broadcast_CreateStreamServer) error {
	ctx := stream.Context()
	md, _ := metadata.FromIncomingContext(ctx)
	values := md["username"]
	username := values[0]

	c.clientsLock.Lock()
	//check client if exist
	user, err := c.userRepo.FindByUsername(username)
	if err != nil {
		return status.Errorf(codes.NotFound, "user not found")
	}

	// add client connection in memory
	c.clients[username] = stream

	// check existing message when client offline
	messages, err := c.messageRepo.ReadByUserId(user.Id)
	if err != nil {
		return status.Errorf(codes.Internal, err.Error())
	}
	// send existing message
	if len(*messages) > 0 {
		var messageId []uint
		for _, v := range *messages {
			messageId = append(messageId, v.Id)

			msg := &chat_server.Message{
				To:        v.To.Username,
				Content:   v.Content,
				CreatedAt: timestamppb.New(v.CreatedAt.Time),
			}
			if err := stream.Send(msg); err != nil {
				if err := c.messageRepo.Create(&v); err != nil {
					log.Printf("Error sending queued message to client %d: %v", v.To.Id, err)
				}
			}
		}
		if err := c.messageRepo.Delete(messageId); err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
	}
	c.clientsLock.Unlock()

	for {
		byteMessage := <-c.consumer.Messages()

		msg := new(domain.KafkaMessage)
		if err := json.Unmarshal(byteMessage.Value, msg); err != nil {
			log.Printf("Error sending queued message to client %s: %v", msg.ToUsername, err)
		}

		if err := c.sendMessageToClient(msg.ToUsername, msg); err != nil {
			//save unsed messsage
			go func(cpMsg *domain.KafkaMessage) {
				userFrom := &domain.User{Id: cpMsg.FromId}
				userTo := &domain.User{Id: cpMsg.ToId}

				message := &domain.Message{Form: userFrom, To: userTo, Content: cpMsg.Content}
				if err := c.messageRepo.Create(message); err != nil {
					log.Printf("Error sending queued message to client %s: %v", cpMsg.ToUsername, err)
				}
			}(msg)
		}
	}
}

func (c *ChatHandler) sendMessageToClient(clientID string, msg *domain.KafkaMessage) error {
	c.clientsLock.Lock()
	defer c.clientsLock.Unlock()
	clientStream, ok := c.clients[clientID]
	if !ok {
		return nil
	}
	message := chat_server.Message{
		To:        msg.ToUsername,
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt,
	}

	if err := clientStream.Send(&message); err != nil {
		return status.Errorf(codes.Internal, err.Error())
	}

	return nil
}

func (c *ChatHandler) BroadcastMessage(ctx context.Context, message *chat_server.Message) (*empty.Empty, error) {
	toUsername := message.GetTo()
	// Extract metadata from the context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("unable to retrieve metadata")
	}

	// Get the 'username' header (or any other header)
	values := md.Get("username")
	if len(values) == 0 {
		return nil, status.Errorf(codes.Aborted, "username not provided in metadata")
	}
	fromUsername := values[0]

	var (
		toUser, fromUser *domain.User
		toErr, fromErr   error
		wg               sync.WaitGroup
	)

	// Use wait group to execute both user lookup calls concurrently
	wg.Add(2)

	go func(cpToUsername string) {
		defer wg.Done()
		toUser, toErr = c.userRepo.FindByUsername(cpToUsername)
	}(toUsername)

	go func(cpFromUsername string) {
		defer wg.Done()
		fromUser, fromErr = c.userRepo.FindByUsername(cpFromUsername)
	}(fromUsername)

	// Wait for both goroutines to complete
	wg.Wait()
	// Handle errors from both goroutines
	if fromErr != nil {
		return nil, status.Errorf(codes.NotFound, "sender user not found")
	}
	if toErr != nil {
		return nil, status.Errorf(codes.NotFound, "recipient user not found")
	}

	// Create Kafka message
	kafkaMessage := &domain.KafkaMessage{
		FromId:     fromUser.Id,
		ToId:       toUser.Id,
		ToUsername: toUser.Username,
		Content:    message.GetContent(),
		CreatedAt:  timestamppb.New(time.Now()),
	}

	// Publish Kafka message in a separate goroutine
	go func() {
		if err := c.publishMessage(c.config.GetString("kafka.topic"), kafkaMessage); err != nil {
			fmt.Printf("error sending message: %v\n", err)
		}
	}()

	return &empty.Empty{}, nil
}

func (c *ChatHandler) publishMessage(topic string, message *domain.KafkaMessage) error {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return status.Errorf(codes.Internal, err.Error())
	}
	_, _, err = c.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(jsonMessage),
	})
	return err
}

func (c *ChatHandler) CreateUser(ctx context.Context, request *chat_server.User) (*empty.Empty, error) {
	user := &domain.User{
		Name:     request.GetName(),
		Username: request.GetUsername(),
	}

	if err := c.userRepo.Create(user); err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	return &empty.Empty{}, nil
}
