package server

import (
	"log"
	"net"

	"github.com/taahakh/crunch/server/talk/talk"
	"google.golang.org/grpc"
)

func Start() {
	listen, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("Failed to listen on port 9000: %v", err)
	}

	talk_server := talk.Server{}

	server := grpc.NewServer()

	talk.RegisterTalkServiceServer(server, &talk_server)

	if err := server.Serve(listen); err != nil {
		log.Fatalf("Failed to serve gRPC server over port 9000")
	}
}
