package models

import (
	"bursa-alert/internal"
	"bursa-alert/lib/utils"
	"sync"
)

type StockEntry struct {
	internal internal.StockEntry
}

func (s StockEntry) ToMap() map[string]float32 {
	return map[string]float32{
		"last_price":            s.GetLastPrice(),
		"preclose_price":        s.GetPreclosePrice(),
		"price_change":          float32(s.GetPriceChange()),
		"total_bought_quantity": float32(s.GetTotalBoughtQuantity()),
		"trade_value":           float32(s.GetTradeValue()),
		"buy_value":             s.GetBuyValue(),
		"buy_volume":            float32(s.GetBuyVolume()),
		"sell_volume":           float32(s.GetSellVolume()),
		"buy_rate":              s.GetBuyRate(),
	}
}

func NewStockEntry(i internal.StockEntry) StockEntry {
	return StockEntry{i}
}

func (s StockEntry) GetIndex() uint {
	return s.internal.StockIndex
}

func (s StockEntry) GetLastPrice() float32 {
	return float32(s.internal.LastPrice) / 1000
}

func (s StockEntry) GetPreclosePrice() float32 {
	return float32(s.internal.PreclosePrice) / 1000
}

func (s StockEntry) GetPriceChange() int {
	return int(s.internal.LastPrice) - int(s.internal.PreclosePrice)
}

func (s StockEntry) GetTotalBoughtQuantity() uint {
	if s.internal.IsPurchase != 1 {
		return 0
	}
	return s.internal.TotalBoughtQuantity * 100
}

func (s StockEntry) GetTradeValue() uint {
	if s.internal.IsPurchase != 1 {
		return 0
	}
	return s.internal.LastPrice / 1000 * s.internal.TotalBoughtQuantity
}

func (s StockEntry) GetBuyValue() float32 {
	if s.internal.BuyValue == 0 {
		return 0
	}
	return float32(s.internal.BuyValue) / 1000
}

func (s StockEntry) GetBuyVolume() uint {
	return s.internal.BuyVolumeMorning + s.internal.BuyVolumeAfternoon
}

func (s StockEntry) GetSellVolume() uint {
	return s.internal.SellVolumeMorning + s.internal.SellVolumeAfternoon
}

func (s StockEntry) GetBuyRate() float32 {
	return float32(s.GetBuyVolume()) / float32(s.GetBuyVolume()+s.GetSellVolume())
}

type StockMetadata struct {
	Name   string
	Ticker string
	Id     int
}

type stockMap struct {
	m map[uint]*StockEntry
	l sync.Mutex
}

func NewStockMap() stockMap {
	return stockMap{
		m: make(map[uint]*StockEntry),
		l: sync.Mutex{},
	}
}

func (s *stockMap) Update(e StockEntry) *StockEntry {
	s.l.Lock()
	defer s.l.Unlock()
	if _, ok := s.m[e.internal.StockIndex]; !ok {
		s.m[e.internal.StockIndex] = &e
		return &e
	}
	tmpPointer := s.m[e.internal.StockIndex]
	if tmpPointer == nil {
		s.m[e.internal.StockIndex] = &e
		return &e
	}
	utils.CopyNonDefaultValues(&e.internal, &tmpPointer.internal)
	return tmpPointer
}
