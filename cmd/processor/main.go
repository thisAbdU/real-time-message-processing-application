package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/thisAbdu/real-time-message-processing-application/internal/models"
)

const movingAverageWindow = 5

var (
	priceHistory = make(map[string][]float64)
	historyMutex sync.Mutex
)

func main() {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: "pulsar://localhost:6650",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            "stock-prices",
		SubscriptionName: "price-processor",
		Type:             pulsar.Shared,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: "processed-stock-data",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	for {
		msg, err := consumer.Receive(context.Background())
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			continue
		}

		var stockPrice models.StockPrice
		err = json.Unmarshal(msg.Payload(), &stockPrice)
		if err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			consumer.Nack(msg)
			continue
		}

		// Process the stock price
		processedData := processStockPrice(stockPrice)

		payload, err := json.Marshal(processedData)
		if err != nil {
			log.Printf("Error marshaling processed data: %v", err)
			consumer.Nack(msg)
			continue
		}

		_, err = producer.Send(context.Background(), &pulsar.ProducerMessage{
			Payload: payload,
		})
		if err != nil {
			log.Printf("Error publishing processed data: %v", err)
			consumer.Nack(msg)
		} else {
			log.Printf("Processed and published data for %s: Price: %.2f, SMA: %.2f", 
				processedData.Symbol, processedData.Price, processedData.MovingAverage)
			consumer.Ack(msg)
		}
	}
}

func processStockPrice(price models.StockPrice) models.ProcessedStockData {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	// Add the current price to the history
	priceHistory[price.Symbol] = append(priceHistory[price.Symbol], price.Price)

	// Keep only the last 'movingAverageWindow' prices
	if len(priceHistory[price.Symbol]) > movingAverageWindow {
		priceHistory[price.Symbol] = priceHistory[price.Symbol][1:]
	}

	// Calculate the simple moving average
	var sum float64
	for _, p := range priceHistory[price.Symbol] {
		sum += p
	}
	movingAverage := sum / float64(len(priceHistory[price.Symbol]))

	return models.ProcessedStockData{
		Symbol:        price.Symbol,
		Price:         price.Price,
		Time:          price.Time,
		MovingAverage: movingAverage,
	}
}
