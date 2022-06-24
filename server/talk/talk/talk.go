package talk

import (
	"context"
	"log"
)

type Server struct {
}

func (s *Server) SayHello(ctx context.Context, t *Talk) (*Talk, error) {
	log.Printf("Recieved message body from client: %s", t.Body)
	return &Talk{Body: "Hello from the server"}, nil
}
