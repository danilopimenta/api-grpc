package main

import (
	"context"
	"log"
	"time"

	"github.com/danilopimenta/api-grpc/pb"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:8082"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewHiClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Saying(ctx, &pb.SayingRequest{Say: "hi"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Say)
}