package models

type StockPrice struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Time   int64   `json:"time"`
}

type ProcessedStockData struct {
	Symbol        string  `json:"symbol"`
	Price         float64 `json:"price"`
	Time          int64   `json:"time"`
	MovingAverage float64 `json:"moving_average"`
}