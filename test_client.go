package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	pb "github.com/thisAbdu/real-time-message-processing-application/api"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewStockServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.GetLatestStockData(ctx, &pb.GetLatestStockDataRequest{Symbol: "AAPL"})
	if err != nil {
		log.Fatalf("could not get stock data: %v", err)
	}
	log.Printf("Stock Data: %s", r.String())

	stream, err := c.StreamStockData(context.Background(), &pb.StreamStockDataRequest{Symbol: "AAPL"})
	if err != nil {
		log.Fatalf("could not stream stock data: %v", err)
	}
	for i := 0; i < 5; i++ {
		r, err := stream.Recv()
		if err != nil {
			log.Fatalf("could not receive: %v", err)
		}
		log.Printf("Streamed Stock Data: %s", r.String())
		time.Sleep(time.Second)
	}
}