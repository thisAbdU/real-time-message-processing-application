package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"github.com/apache/pulsar-client-go/pulsar"
	_ "github.com/lib/pq"
	"github.com/thisAbdu/real-time-message-processing-application/internal/models"
)

func main() {
	pulsarClient, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: os.Getenv("PULSAR_URL"),
	})
	if err != nil {
		log.Fatalf("Could not create Pulsar client: %v", err)
	}
	defer pulsarClient.Close()
	
	// Initialize PostgreSQL connection
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	// Create the stocks table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS stocks (
			symbol VARCHAR(10) NOT NULL,
			price DECIMAL(10, 2) NOT NULL,
			time BIGINT NOT NULL,
			moving_average DECIMAL(10, 2) NOT NULL,
			PRIMARY KEY (symbol, time)
		)
	`)
	if err != nil {
		log.Fatalf("Could not create stocks table: %v", err)
	}

	// Subscribe to the processed-stock-data topic
	consumer, err := pulsarClient.Subscribe(pulsar.ConsumerOptions{
		Topic:            "processed-stock-data",
		SubscriptionName: "storage-service",
	})
	if err != nil {
		log.Fatalf("Could not subscribe to topic: %v", err)
	}
	defer consumer.Close()

	for {
		msg, err := consumer.Receive(context.Background())
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			continue
		}

		var stockData models.ProcessedStockData
		err = json.Unmarshal(msg.Payload(), &stockData)
		if err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			consumer.Nack(msg)
			continue
		}

		// Insert the stock data into the database
		_, err = db.Exec(`
			INSERT INTO stocks (symbol, price, time, moving_average)
			VALUES ($1, $2, $3, $4)
		`, stockData.Symbol, stockData.Price, stockData.Time, stockData.MovingAverage)
		if err != nil {
			log.Printf("Error inserting stock data: %v", err)
			consumer.Nack(msg)
			continue
		}

		consumer.Ack(msg)
		log.Printf("Stored stock data for %s", stockData.Symbol)
	}
}
