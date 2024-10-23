package main

import (
	"context"
	"log"

	proto "chitchat"

	"google.golang.org/grpc"
)

func main() {
	connection, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer connection.Close()

	client := proto.NewChitChatClient(connection)

	//Join the chat
	joinResponse, err := client.Join(context.Background(), &pb.JoinRequest{ParticipantId: "User1"})
	if err != nil {
		log.Fatalf("could not join: %v", err)
	}
	log.Printf("Join response: %s", joinResponse.WelcomeMessage)

	//Publish a message

	//Leave the chat
}
