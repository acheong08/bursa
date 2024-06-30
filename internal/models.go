package internal

type StockEntry struct {
	StockIndex uint `json:"1"`
	// -- MT Data
	// In milicents.s. Divide by 1000 to get ringgit
	LastPrice uint `json:"209"`
	// In milicents
	PreclosePrice uint `json:"153"`
	// 1 - true
	IsPurchase uint `json:"248"`
	// Should only exist if is purchase
	TotalBoughtQuantity uint `json:"210"`

	// -- SM Data
	AccumulatedValue float32 `json:"90"`
	// In milicents
	BuyValue            uint `json:"87"`
	BuyVolumeMorning    uint `json:"114"`
	BuyVolumeAfternoon  uint `json:"120"`
	SellVolumeMorning   uint `json:"117"`
	SellVolumeAfternoon uint `json:"123"`
}
