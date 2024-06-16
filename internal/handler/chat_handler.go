package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/IBM/sarama"
	"github.com/PNYwise/chat-server/internal/domain"
	chat_server "github.com/PNYwise/chat-server/proto"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChatHandler struct {
	clients      map[string]chat_server.Broadcast_CreateStreamServer
	clientsLock  sync.Mutex
	messageQueue chan *chat_server.Message
	config       *viper.Viper
	producer     sarama.SyncProducer
	consumer     sarama.PartitionConsumer
	userRepo     domain.IUserRepository
	messageRepo  domain.IMessageRepository
	chat_server.UnimplementedBroadcastServer
}

func NewChatHandler(config *viper.Viper, producer sarama.SyncProducer, consumer sarama.PartitionConsumer, userRepo domain.IUserRepository, messageRepo domain.IMessageRepository) *ChatHandler {
	return &ChatHandler{
		clients:      make(map[string]chat_server.Broadcast_CreateStreamServer),
		messageQueue: make(chan *chat_server.Message),
		config:       config,
		producer:     producer,
		consumer:     consumer,
		userRepo:     userRepo,
		messageRepo:  messageRepo,
	}
}

func (c *ChatHandler) CreateStream(connect *chat_server.Connect, stream chat_server.Broadcast_CreateStreamServer) error {
	stringClientID := connect.GetId()

	clientID, err := strconv.Atoi(stringClientID)
	if err != nil {
		return status.Errorf(codes.Internal, err.Error())
	}
	c.clientsLock.Lock()
	//check client if exist, if no insert new client
	if ok := c.userRepo.Exist(clientID); !ok {
		return status.Errorf(codes.NotFound, "user not found")
	}

	// add client connection in memory
	c.clients[stringClientID] = stream

	// check existing message when client offline
	messages, err := c.messageRepo.ReadByUserId(uint(clientID))
	if err != nil {
		return status.Errorf(codes.Internal, err.Error())
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
			return status.Errorf(codes.Internal, err.Error())
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

				userFrom := &domain.User{Id: uint(from)}
				userTo := &domain.User{Id: uint(to)}

				message := &domain.Message{Form: userFrom, To: userTo, Content: cpMsg.GetContent()}
				if err := c.messageRepo.Create(message); err != nil {
					log.Printf("Error sending queued message to client %s: %v", cpMsg.GetTo(), err)
				}
			}(msg)
		}
	}
}

func (c *ChatHandler) sendMessageToClient(clientID string, msg *chat_server.Message) error {
	c.clientsLock.Lock()
	defer c.clientsLock.Unlock()
	clientStream, ok := c.clients[clientID]
	if !ok {
		return nil
	}

	if err := clientStream.Send(msg); err != nil {
		return status.Errorf(codes.Internal, err.Error())
	}

	return nil
}

func (c *ChatHandler) BroadcastMessage(ctx context.Context, message *chat_server.Message) (*chat_server.Close, error) {
	toId, err := strconv.Atoi(message.GetTo())
	if err != nil {
		return nil, err
	}
	exist := c.userRepo.Exist(toId)
	if !exist {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}
	go func() {
		if err := c.publishMessage(c.config.GetString("kafka.topic"), message); err != nil {
			fmt.Printf("err send message :%v", err)
		}
	}()
	return &chat_server.Close{}, nil
}

func (c *ChatHandler) publishMessage(topic string, message *chat_server.Message) error {
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
