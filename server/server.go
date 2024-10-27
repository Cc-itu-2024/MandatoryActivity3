package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	pb "chitchat/chitchat"

	"google.golang.org/grpc"
)

type ChitChatServer struct {
	pb.UnimplementedChitChatServer
	participants      map[string]bool
	streams           map[string]chan *pb.BroadcastMessage
	mu                sync.Mutex
	lamportTime       int64
	nextParticipantId int
}

func NewChitChatServer() *ChitChatServer {
	return &ChitChatServer{
		participants: make(map[string]bool),
		streams:      make(map[string]chan *pb.BroadcastMessage),
	}
}

func (s *ChitChatServer) Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	participantId := fmt.Sprintf("Participant %d", s.nextParticipantId+1)
	s.nextParticipantId++

	s.participants[participantId] = true
	s.streams[participantId] = make(chan *pb.BroadcastMessage, 10)

	s.lamportTime++
	joinMessage := &pb.BroadcastMessage{
		Message:       fmt.Sprintf("Participant %s joined at Lamport time %d", participantId, s.lamportTime),
		Time:          s.lamportTime,
		ParticipantId: participantId,
	}
	go s.sendToAllStreams(joinMessage, participantId)

	welcomeMessage := fmt.Sprintf("Welcome to ChitChat, %s!", participantId)
	log.Printf("%s joined the chat at Lamport time %d", participantId, s.lamportTime)
	return &pb.JoinResponse{Hello: welcomeMessage, ParticipantId: participantId}, nil
}

func (s *ChitChatServer) Leave(ctx context.Context, req *pb.LeaveRequest) (*pb.LeaveResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	participantId := req.ParticipantId
	delete(s.participants, participantId)
	delete(s.streams, participantId)

	s.lamportTime++
	leaveMessage := &pb.BroadcastMessage{
		Message:       fmt.Sprintf("Participant %s left at Lamport time %d", participantId, s.lamportTime),
		Time:          s.lamportTime,
		ParticipantId: participantId,
	}
	go s.sendToAllStreams(leaveMessage, participantId)

	log.Printf("%s left the chat at Lamport time %d", participantId, s.lamportTime)
	return &pb.LeaveResponse{ByeMessage: "Goodbye!"}, nil
}

func (s *ChitChatServer) sendToAllStreams(msg *pb.BroadcastMessage, excludeId string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for participantId, stream := range s.streams {
		if participantId != excludeId{
		select {
		case stream <- msg:
		default:
			log.Printf("Failed to send message to %s: channel is full", participantId)
		}
	}
}
}

func (s *ChitChatServer) PublishMessage(ctx context.Context, req *pb.ChatMessage) (*pb.PublishResponse, error) {
	if len(req.Message) > 128 {
		return nil, fmt.Errorf("message exceeds maximum length of 128 characters")
	}

	s.mu.Lock()
	s.lamportTime++
	messageTime := s.lamportTime
	s.mu.Unlock()

	msg := fmt.Sprintf("Message from %s: %s", req.ParticipantId, req.Message)
	s.sendToAllStreams(&pb.BroadcastMessage{
		Message:       msg,
		Time:          messageTime,
		ParticipantId: req.ParticipantId,
	}, req.ParticipantId)

	log.Printf("%s published message at Lamport time %d: %s", req.ParticipantId, messageTime, req.Message)
	return &pb.PublishResponse{Status: "Message delivered"}, nil
}

func (s *ChitChatServer) ReceiveMessages(req *pb.JoinRequest, stream pb.ChitChat_ReceiveMessagesServer) error {
	s.mu.Lock()
	participantId := req.ParticipantId
	if _, exists := s.participants[participantId]; !exists {
		s.mu.Unlock()
		return fmt.Errorf("participant %s not found", participantId)
	}

	streamChannel := s.streams[participantId]
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.streams, participantId)
		s.mu.Unlock()
	}()

	for {
		msg, ok := <-streamChannel
		if !ok {
			return nil
		}
		if err := stream.Send(&pb.BroadcastNotification{Message: msg}); err != nil {
			return err
		}
	}
}

func main() {
	logFile, err := os.OpenFile("chitchat.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChitChatServer(grpcServer, NewChitChatServer())

	log.Printf("Server listening on %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
