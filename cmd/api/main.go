package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/thisAbdu/real-time-message-processing-application/api"
)

type server struct {
	pb.UnimplementedStockServiceServer
	db *sql.DB
}

func (s *server) GetLatestStockData(ctx context.Context, req *pb.GetLatestStockDataRequest) (*pb.StockData, error) {
	var stockData pb.StockData
	err := s.db.QueryRowContext(ctx, `
		SELECT symbol, price, time, moving_average
		FROM stocks
		WHERE symbol = $1
		ORDER BY time DESC
		LIMIT 1
	`, req.Symbol).Scan(&stockData.Symbol, &stockData.Price, &stockData.Time, &stockData.MovingAverage)
	if err != nil {
		return nil, err
	}
	return &stockData, nil
}

func (s *server) StreamStockData(req *pb.StreamStockDataRequest, stream pb.StockService_StreamStockDataServer) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stockData, err := s.GetLatestStockData(stream.Context(), &pb.GetLatestStockDataRequest{Symbol: req.Symbol})
			if err != nil {
				return err
			}
			if err := stream.Send(stockData); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func main() {
	// Connect to the database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create a gRPC server
	s := grpc.NewServer()
	pb.RegisterStockServiceServer(s, &server{db: db})
	
	// Enable reflection for easier debugging
	reflection.Register(s)

	// Start listening
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Println("gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
