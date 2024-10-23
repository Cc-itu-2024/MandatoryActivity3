package server

import (
	proto "chitchat/chitchat"
	//"context"
	//"log"
	//"net"
	//"google.golang.org/grpc"
)

type ClientInformation struct {
	proto.UnimplementedChitChatServer
	stream proto.ChitChat_CreateStreamServer
	id     string
	active bool
	//name   string
	errorChannel chan error
}

type Server struct {
	proto.UnimplementedChitChatServer
	//timestamp int64
	ConnectedClients []*ClientInformation //Vi får en liste af de clients, der connected til serveren
}

func (p *Server) CreateStream(participant *proto.Participant, stream proto.ChitChat_CreateStreamServer) error { //error er returværdien
	clientInformation := &ClientInformation{
		stream: stream,
		id:     participant.Id,
		active: true,
		errorChannel:  make(chan error), //laver en channel der modtager en error, når der skal lukkes ned for forbindelsen
	}

	p.ConnectedClients = append(p.ConnectedClients, clientInformation) //connected clients bliver tilføjet til arrayet

	return <-clientInformation.errorChannel
}

func BroadCastMessage() {

}

func removeDisconnectedClients() {

}

func main() {

}
