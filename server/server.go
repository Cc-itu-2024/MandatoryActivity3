package server

import (
	proto "chitchat/chitchat"
	//"context"
	//"log"
	//"net"
	//"google.golang.org/grpc"
)

type Connection struct {
	proto.UnimplementedBroadcastServer
	stream proto.Broadcast_CreateStreamServer
	id     string
	active bool
	//name   string
	error chan error
}

type Server struct {
	proto.UnimplementedBroadcastServer
	//timestamp int64
	Connection []*Connection
}

func (p *Server) CreateStream(pconn *proto.Connect, stream proto.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream: stream,
		id:     pconn.Participant.Id,
		active: true,
		error:  make(chan error),
	}

	p.Connection = append(p.Connection, conn)

	return <-conn.error
}

func BroadCastMessage() {

}

func removeDisconnectedClients() {

}

func main() {

}
