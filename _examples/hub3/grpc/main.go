package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/delving/hub3/pkg/server/grpc/pb/namespacepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	address     = "localhost:3001"
	defaultName = "http://purl.org/dc/elements/1.1/title"
)

func main() {
	var conn *grpc.ClientConn
	var err error
	// Create the client TLS credentials
	creds, err := credentials.NewClientTLSFromFile("../../certs/server.crt", "")
	if err != nil {
		log.Fatalf("could not load tls cert: %s", err)
	}

	// Set up a connection to the server.
	conn, err = grpc.Dial(
		address,
		//grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewNamespaceClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SearchLabel(ctx, &pb.SearchLabelRequest{Uri: name})
	if err != nil {
		log.Fatalf("could not find namespace: %v", err)
	}
	log.Printf("namespace: %s", r.Label)
}
