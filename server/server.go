package main

import (
	"context"
	"fmt"
	"log"
	"net"
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

	// Assign the next participant ID
	participantId := fmt.Sprintf("User%d", s.nextParticipantId)
	s.nextParticipantId++

	// Store participant and create a channel for updates
	s.participants[participantId] = true
	s.streams[participantId] = make(chan *pb.BroadcastMessage, 10) // Buffered channel

	// Broadcast that a new participant has joined
	s.lamportTime++
	joinMessage := &pb.BroadcastMessage{
		Message:       fmt.Sprintf("Participant %s joined at Lamport time %d", participantId, s.lamportTime),
		Time:          s.lamportTime,
		ParticipantId: participantId,
	}
	go s.sendToAllStreams(joinMessage)

	welcomeMessage := fmt.Sprintf("Welcome to ChitChat, %s!", participantId)
	log.Printf("Participant %s joined the chat at Lamport time %d", participantId, s.lamportTime)
	return &pb.JoinResponse{Hello: welcomeMessage}, nil
}

func (s *ChitChatServer) Leave(ctx context.Context, req *pb.LeaveRequest) (*pb.LeaveResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	participantId := req.ParticipantId
	delete(s.participants, participantId)
	delete(s.streams, participantId)

	// Broadcast that a participant has left
	s.lamportTime++
	leaveMessage := &pb.BroadcastMessage{
		Message:       fmt.Sprintf("Participant %s left at Lamport time %d", participantId, s.lamportTime),
		Time:          s.lamportTime,
		ParticipantId: participantId,
	}
	go s.sendToAllStreams(leaveMessage)

	log.Printf("Participant %s left the chat at Lamport time %d", participantId, s.lamportTime)
	return &pb.LeaveResponse{ByeMessage: "Goodbye!"}, nil
}

// Broadcast a message to all client streams
func (s *ChitChatServer) sendToAllStreams(msg *pb.BroadcastMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for participantId, stream := range s.streams {
		select {
		case stream <- msg:
			log.Printf("Broadcasting message to %s: %s", participantId, msg.Message)
		default:
			log.Printf("Failed to send message to %s: channel is full", participantId)
		}
	}
}

func (s *ChitChatServer) PublishMessage(ctx context.Context, req *pb.ChatMessage) (*pb.PublishResponse, error) {
	s.mu.Lock()
	s.lamportTime++
	messageTime := s.lamportTime
	s.mu.Unlock()

	// Broadcast the message to all participants
	msg := fmt.Sprintf("Message from %s: %s", req.ParticipantId, req.Id)
	s.sendToAllStreams(&pb.BroadcastMessage{
		Message:       msg,
		Time:          messageTime,
		ParticipantId: req.ParticipantId,
	})

	log.Printf("Published message from %s at Lamport time %d: %s", req.ParticipantId, messageTime, req.Id)
	return &pb.PublishResponse{Status: "Message delivered"}, nil
}

func (s *ChitChatServer) ReceiveMessages(req *pb.JoinRequest, stream pb.ChitChat_ReceiveMessagesServer) error {
	s.mu.Lock()
	participantId := req.ParticipantId
	if _, exists := s.participants[participantId]; !exists {
		s.mu.Unlock()
		return fmt.Errorf("participant %s not found", participantId)
	}
	// We already have a stream for this participant; no need to create a new one
	streamChannel := s.streams[participantId]
	s.mu.Unlock()

	// Defer cleanup when the stream is closed
	defer func() {
		s.mu.Lock()
		delete(s.streams, participantId)
		s.mu.Unlock()
	}()

	// Listen for messages and send them to the client
	for {
		msg, ok := <-streamChannel
		if !ok {
			return nil // Channel closed, exit
		}
		if err := stream.Send(&pb.BroadcastNotification{Message: msg}); err != nil {
			return err
		}
	}
}

func main() {
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
