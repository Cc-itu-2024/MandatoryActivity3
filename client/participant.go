package main

import (
	"bufio"
	"context"
	"log"
	"os"

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

	joinResp, err := client.Join(context.Background(), &pb.JoinRequest{ParticipantId: ""})
	if err != nil {
		log.Fatalf("could not join: %v", err)
	}
	log.Printf("Join response: %s", joinResp.Hello)

	participantId := joinResp.ParticipantId

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
			log.Printf("Received: %s at Lamport time %d", notification.Message.Message, notification.Message.Time)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		log.Print("Enter your message or write exit to leave")
		if scanner.Scan() {
			message := scanner.Text()

			// Send meddelelsen til chatten
			pubResp, err := client.PublishMessage(context.Background(), &pb.ChatMessage{
				Message:       message,
				ParticipantId: participantId,
			})
			if err != nil {
				log.Fatalf("could not publish: %v", err)
			}
			if message == "exit" {
				LeaveChat(client, participantId) // der skal ske noget her
				break
			}
			log.Printf("Publish response: %s", pubResp.Status)
		}

		/*
			pubResp, err := client.PublishMessage(context.Background(), &pb.ChatMessage{
				Message:            "Hello, ChitChat, I am here!",
				ParticipantId: participantId,
			})
			if err != nil {
				log.Fatalf("could not publish: %v", err)
			}
			log.Printf("Publish response: %s", pubResp.Status)
		*/

		/*
				time.Sleep(3 * time.Second)

				leaveResp, err := client.Leave(context.Background(), &pb.LeaveRequest{
					ParticipantId: participantId,
				})
				if err != nil {
					log.Fatalf("could not leave: %v", err)
				}
				log.Printf("Leave response: %s", leaveResp.ByeMessage)
			}
		*/
	}
}
func LeaveChat(client pb.ChitChatClient, participantId string) {
	leaveResp, err := client.Leave(context.Background(), &pb.LeaveRequest{
		ParticipantId: participantId,
	})
	if err != nil {
		log.Fatal("failed to leave chat :(")
	}

	log.Printf("%s %s %d", participantId, leaveResp.ByeMessage)
}
