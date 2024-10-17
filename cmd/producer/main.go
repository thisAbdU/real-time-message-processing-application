package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/thisAbdu/real-time-message-processing-application/internal/models"
)

func main() {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: os.Getenv("PULSAR_URL"),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: "stock-prices",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	symbols := []string{"AAPL", "GOOGL", "MSFT", "AMZN"}

	for {
		symbol := symbols[rand.Intn(len(symbols))]
		price := 100 + rand.Float64()*900
		stockPrice := models.StockPrice{
			Symbol: symbol,
			Price:  price,
			Time:   time.Now().Unix(),
		}

		payload, err := json.Marshal(stockPrice)
		if err != nil {
			log.Printf("Failed to marshal stock price: %v", err)
			continue
		}

		_, err = producer.Send(context.Background(), &pulsar.ProducerMessage{
			Payload: payload,
		})
		if err != nil {
			log.Printf("Failed to publish message: %v", err)
		} else {
			log.Printf("Published stock price: %s $%.2f", symbol, price)
		}

		time.Sleep(time.Second)
	}
}
