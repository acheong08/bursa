package alerts

import (
	"bursa-alert/lib/models"
	"fmt"
	"strconv"
)

type Alert struct {
	Label string   `yaml:"label" json:"label"`
	Rules []Rule   `yaml:"rules" json:"rules"`
	Tags  []string `yaml:"tags" json:"tags"`
}

type Rule struct {
	A   eval       `yaml:"a" json:"a"`
	Cmp comparator `yaml:"cmp" json:"cmp"`
	B   eval       `yaml:"b" json:"b"`
}
type eval struct {
	Type  valueType `yaml:"type" json:"type"`
	Value variable  `yaml:"value" json:"value"`
}

type (
	variable   string
	comparator string
	valueType  string
)

const (
	CmpEquals comparator = "=="
	CmpNot    comparator = "!="
	CmpGt     comparator = ">"
	CmpGte    comparator = ">="
	CmpLt     comparator = "<"
	CmpLte    comparator = "<="
)

const (
	EvalVariable valueType = "var"
	EvalConstant valueType = "const"
)

const (
	VarLastPrice           variable = "last_price"
	VarPreclosePrice       variable = "preclose_price"
	VarPriceChange         variable = "price_change"
	VarTotalBoughtQuantity variable = "total_bought_quantity"
	VarTradeValue          variable = "trade_value"
	VarBuyVolume           variable = "buy_volume"
	VarSellVolume          variable = "sell_volume"
	VarBuyRate             variable = "buy_rate"
)

func (a Alert) Validate() error {
	for _, rule := range a.Rules {
		if err := rule.validate(); err != nil {
			return fmt.Errorf("alert %s failed to parse: %w", a.Label, err)
		}
	}
	return nil
}

func (e eval) validate() error {
	switch e.Type {
	default:
		return fmt.Errorf("%s is not a valid type", e.Type)
	case EvalConstant, EvalVariable:
	}
	if e.Type == EvalVariable {
		switch e.Value {
		default:
			return fmt.Errorf("%s is not a valid variable", e.Value)
		case VarLastPrice, VarPreclosePrice, VarPriceChange, VarTotalBoughtQuantity, VarTradeValue, VarBuyVolume, VarSellVolume, VarBuyRate:
		}
	} else {
		if _, err := strconv.ParseFloat(string(e.Value), 32); err != nil {
			return fmt.Errorf("%s is not a float", e.Value)
		}
	}
	return nil
}

func (e *eval) float32(se models.StockEntry) float32 {
	if e.Type == EvalConstant {
		f, err := strconv.ParseFloat(string(e.Value), 32)
		if err != nil {
			panic(err)
		}
		return float32(f)
	}
	switch e.Value {
	case VarLastPrice:
		return se.GetLastPrice()
	case VarPreclosePrice:
		return se.GetPreclosePrice()
	case VarPriceChange:
		return float32(se.GetPriceChange())
	case VarTotalBoughtQuantity:
		return float32(se.GetTotalBoughtQuantity())
	case VarTradeValue:
		return float32(se.GetTradeValue())
	case VarBuyVolume:
		return float32(se.GetBuyVolume())
	case VarSellVolume:
		return float32(se.GetSellVolume())
	case VarBuyRate:
		return se.GetBuyRate()
	default:
		panic("invalid variable")
	}
}

func (r Rule) validate() error {
	if err := r.A.validate(); err != nil {
		return fmt.Errorf("Rule %s %s %s failed to parse: %w", r.A, r.Cmp, r.B, err)
	}
	if err := r.B.validate(); err != nil {
		return fmt.Errorf("Rule %s %s %s failed to parse: %w", r.A, r.Cmp, r.B, err)
	}
	switch r.Cmp {
	default:
		return fmt.Errorf("%s is not a valid comparator", r.Cmp)
	case CmpEquals, CmpNot, CmpGt, CmpGte, CmpLt, CmpLte:
	}
	return nil
}

func (r *Rule) eval(se models.StockEntry) bool {
	a := r.A.float32(se)
	b := r.B.float32(se)
	switch r.Cmp {
	case CmpEquals:
		return a == b
	case CmpNot:
		return a != b
	case CmpGt:
		return a > b
	case CmpGte:
		return a >= b
	case CmpLt:
		return a < b
	case CmpLte:
		return a <= b
	default:
		panic("invalid comparator")
	}
}

func (a Alert) Eval(se models.StockEntry) bool {
	for _, rule := range a.Rules {
		if !rule.eval(se) {
			return false
		}
	}
	return true
}
