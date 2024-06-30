package alerts_test

import (
	"bursa-alert/lib/alerts"
	"testing"

	"gopkg.in/yaml.v3"
)

const VALID = `
- label: "Price Increase Alert"
  rules:
    - a:
        type: "var"
        value: "last_price"
      cmp: ">"
      b:
        type: "var"
        value: "preclose_price"
  tags:
    - "price"
    - "increase"

- label: "High Volume Alert"
  rules:
    - a:
        type: "var"
        value: "total_bought_quantity"
      cmp: ">"
      b:
        type: "const"
        value: "1000000"
    - a:
        type: "var"
        value: "buy_rate"
      cmp: ">="
      b:
        type: "const"
        value: "0.6"
  tags:
    - "volume"
    - "high"
`

const INVALID = `
- label: "Invalid Comparator"
  rules:
    - a:
        type: "var"
        value: "last_price"
      cmp: "!=="  # Invalid comparator
      b:
        type: "const"
        value: "100"
  tags:
    - "invalid"

- label: "Invalid Value Type"
  rules:
    - a:
        type: "variable"  # Invalid type, should be "var" or "const"
        value: "last_price"
      cmp: ">"
      b:
        type: "const"
        value: "100"
  tags:
    - "invalid"

- label: "Invalid Variable"
  rules:
    - a:
        type: "var"
        value: "current_price"  # Invalid variable name
      cmp: ">"
      b:
        type: "const"
        value: "100"
  tags:
    - "invalid"

- label: "Invalid Constant"
  rules:
    - a:
        type: "var"
        value: "last_price"
      cmp: ">"
      b:
        type: "const"
        value: "not a number"  # Should be a float32
  tags:
    - "invalid"
`

func TestValid(t *testing.T) {
	var a []alerts.Alert = make([]alerts.Alert, 0)
	if err := yaml.Unmarshal([]byte(VALID), &a); err != nil {
		panic(err)
	}
	for _, alert := range a {
		if err := alert.Validate(); err != nil {
			t.Error(err)
		}
	}
}

func TestInvalid(t *testing.T) {
	var a []alerts.Alert = make([]alerts.Alert, 0)
	if err := yaml.Unmarshal([]byte(INVALID), &a); err != nil {
		panic(err)
	}
	for _, alert := range a {
		if err := alert.Validate(); err == nil {
			t.Errorf("Alert %s is valid", alert.Label)
		}
	}
}
