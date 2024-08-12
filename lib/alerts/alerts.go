package alerts

import (
	"bursa-alert/lib/global"
	"bursa-alert/lib/models"
	"fmt"
	"strconv"
	"time"
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
	Var   variable  `yaml:"variable" json:"variable"`
	Const float32   `yaml:"constant" json:"constant"`
}

type variable struct {
	T        variableType  `yaml:"type" json:"type"`
	// Oldest entry within x minutes
	D uint `yaml:"duration" json:"duration"`
}

type (
	variableType string
	comparator   string
	valueType    string
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
	VarLastPrice           variableType = "last_price"
	VarPreclosePrice       variableType = "preclose_price"
	VarPriceChange         variableType = "price_change"
	VarTotalBoughtQuantity variableType = "total_bought_quantity"
	VarTradeValue          variableType = "trade_value"
	VarBuyVolume           variableType = "buy_volume"
	VarSellVolume          variableType = "sell_volume"
	VarBuyRate             variableType = "buy_rate"
)

func (a Alert) Validate() error {
	for _, rule := range a.Rules {
		if err := rule.validate(); err != nil {
			return fmt.Errorf("alert %s failed to parse: %w", a.Label, err)
		}
	}
	return nil
}

func (e eval) String() string {
	if e.Type == EvalConstant {
		return fmt.Sprintf("%3f", e.Const)
	}
	return fmt.Sprintf("%s (%d min)", string(e.Var.T), e.Var.D)
}

func (e eval) validate() error {
	switch e.Type {
	default:
		return fmt.Errorf("%s is not a valid type", e.Type)
	case EvalConstant, EvalVariable:
	}
	if e.Type == EvalVariable {
		switch e.Var.T {
		default:
			return fmt.Errorf("%s is not a valid variable", e.Var.T)
		case VarLastPrice, VarPreclosePrice, VarPriceChange, VarTotalBoughtQuantity, VarTradeValue, VarBuyVolume, VarSellVolume, VarBuyRate:
		}
	} else {
		if _, err := strconv.ParseFloat(string(e.Var.T), 32); err != nil {
			return fmt.Errorf("%s is not a float", e.Var.T)
		}
	}
	return nil
}

func (e *eval) float32(id uint) float32 {
	if e.Type == EvalConstant {
		f, err := strconv.ParseFloat(string(e.Var.T), 32)
		if err != nil {
			panic(err)
		}
		return float32(f)
	}
	var se models.StockEntry
	if e.Var.D == 0 {
		if e := global.Entries.FetchOne(id); e != nil {
			se = *e
		} else {
			return 0
		}
	} else {
		if e := global.Entries.FetchOldestWithin(id, time.Duration(e.Var.D) * time.Minute); e != nil {
			se = *e
		} else {
			return 0
		}
	}
	switch e.Var.T {
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

func (r *Rule) eval(id uint) bool {
	a := r.A.float32(id)
	b := r.B.float32(id)
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

func (a Alert) Eval(id uint) bool {
	for _, rule := range a.Rules {
		if !rule.eval(id) {
			return false
		}
	}
	return true
}
