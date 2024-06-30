// Default handlers
package handlers

import (
	"bursa-alert/internal/ws"
	"bursa-alert/lib/models"
	"encoding/json"
)

type slData struct {
	Names   []string `json:"77"`
	Tickers []string `json:"2"`
	Offset  uint     `json:"4"`
}

func GetStockMapping(si map[uint]models.StockMetadata) ws.MessageHandler {
	return func(_ *ws.Connection, b []byte) error {
		var sl slData
		if err := json.Unmarshal(b, &sl); err != nil {
			return err
		}
		for i, ticker := range sl.Tickers {
			si[uint(i)+sl.Offset] = models.StockMetadata{
				Name:   sl.Names[i],
				Ticker: ticker,
				Id:     i + int(sl.Offset),
			}
		}
		return nil
	}
}
