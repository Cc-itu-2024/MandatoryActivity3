package server

import (
	proto "chitchat"
	"context"
)

// serveren
type ChitChatServer struct {
	proto.UnimplementedBroadcastServer
	participants map[string]int64
	id           string
	active       bool
	//name   string
	error chan error
}

func NewChitChatServer() *ChitChatServer {
	return &ChitChatServer{
		participants: make(map[string]int64),
	}
}

//Der er en funktion til hver handling i programmet.
//I parameteren tager de messagesne fra chitchat.proto.
//Inde i disse metoder skal der så oprettes logik for hvad der sker når
// hver af de her "services" skal eksekveres.

// functions til at Joine en chat
func (s *ChitChatServer) Join(ctx context.Context, req *proto.JoinRequest) (*proto.JoinResponse, error) {

}

// funtions til at forlade en chat
func (s *ChitChatServer) Leave(ctx context.Context, req *proto.LeaveRequest) (*proto.LeaveResponse, error) {
}

// funtion til at publicere en besked
func (s *ChitChatServer) PublishMessage(ctx context.Context, req *proto.ChatMessage) (*proto.PublishResponse, error) {

}

//funktion der skal broadcaste besked

func (s *ChitChatServer) Broadcast(ctx context.Context, req *proto.BroadcastMessage) (*proto.BroadcastResponse, error) {

}

func main() {

}
