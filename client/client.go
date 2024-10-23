package main

import (
	"context"
	"flag"
	"io"
	"log"
	"time"

	proto "chitchat/chitchat"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	name = flag.String("name", "Anonymous", "Name to greet")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := proto.NewChitChatClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	log.Printf("Name: %s", *name)
	messageStream, err := c.CreateStream(ctx, &proto.Participant{Name: *name})
	if err != nil {
		log.Fatalf("client.ListFeatures failed: %v", err)
	}
	for {
		message, err := messageStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("client.ListFeatures failed: %v", err)
		}
		log.Printf(message.Content)
	}

	
}
