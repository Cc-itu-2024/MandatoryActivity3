package main

import (
	"context"
	"log"
	"time"

	pb "chitchat/chitchat"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewChitChatClient(conn)

	// Join the chat
	joinResp, err := client.Join(context.Background(), &pb.JoinRequest{})
	if err != nil {
		log.Fatalf("could not join: %v", err)
	}
	log.Printf("Join response: %s", joinResp.Hello)

	// Save the assigned ParticipantId from Join response
	participantId := joinResp.Hello

	// Listen for incoming messages
	go func() {
		stream, err := client.ReceiveMessages(context.Background(), &pb.JoinRequest{ParticipantId: participantId})
		if err != nil {
			log.Fatalf("could not receive messages: %v", err)
		}
		for {
			notification, err := stream.Recv()
			if err != nil {
				log.Fatalf("Error receiving message: %v", err)
			}
			log.Printf("Received: %s at time %d", notification.Message.Message, notification.Message.Time)
		}
	}()

	// Publish a message
	pubResp, err := client.PublishMessage(context.Background(), &pb.ChatMessage{
		Id:            "Hello, ChitChat, I am here!", // Used as message content, not as ID
		ParticipantId: participantId, // Use assigned ID from server
	})
	if err != nil {
		log.Fatalf("could not publish: %v", err)
	}
	log.Printf("Publish response: %s", pubResp.Status)

	// Wait before ending
	time.Sleep(3 * time.Second)

	// Send a Leave request
	leaveResp, err := client.Leave(context.Background(), &pb.LeaveRequest{
		ParticipantId: participantId, // Use assigned ID from server
	})
	if err != nil {
		log.Fatalf("could not leave: %v", err)
	}
	log.Printf("Leave response: %s", leaveResp.ByeMessage)
}
