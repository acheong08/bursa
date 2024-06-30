package lib

import (
	"bursa-alert/internal"
	"bursa-alert/internal/ws"
	"bursa-alert/lib/handlers"
	"bursa-alert/lib/models"
	"context"
	"encoding/json"
	"log"
)

func GetDataStream(ctx context.Context, subscriptions []uint, ch chan models.StockEntry, options ...ws.OptionModifier) error {
	if options == nil {
		options = make([]ws.OptionModifier, 0)
	}
	for {
		newCtx, cancel := context.WithCancel(ctx)
		select {
		case <-ctx.Done():
			cancel()
			return nil
		default:
			conn, err := ws.NewConnection(newCtx, subscriptions, options...)
			if err != nil {
				cancel()
				log.Println(err)
				continue
			}
			stockHandler := func(_ *ws.Connection, b []byte) error {
				var internalStockEntry internal.StockEntry
				if err := json.Unmarshal(b, &internalStockEntry); err != nil {
					log.Println(err)
					return err
				}
				ch <- models.NewStockEntry(internalStockEntry)
				return nil
			}
			conn.AddHandler("MT", stockHandler)
			conn.AddHandler("SM", stockHandler)
			if err := conn.StartReadLoop(); err != nil {
				cancel()
				log.Println(err)
				continue
			}
			cancel()
		}
	}
}

func GetStockMetadata(m map[uint]models.StockMetadata) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := ws.NewConnection(ctx, []uint{}, ws.WithMessageHandler("SL", handlers.GetStockMapping(m)))
	return err
}
