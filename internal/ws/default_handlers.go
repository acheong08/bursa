package ws

import "time"

func pingHandler(c *Connection, _ []byte) error {
	if c.timeout != nil {
		c.timeout.Reset(30 * time.Second)
	}

	return c.WriteJson(map[string]any{"mt": "AC", "data": map[string]any{}})
}

func v3RdHandler(c *Connection, _ []byte) error {
	if err := c.WriteJson(rfData(0)); err != nil {
		return nil
	}
	if err := c.WriteJson(rfData(1)); err != nil {
		return nil
	}
	if err := c.WriteJson(rfData(1)); err != nil {
		return nil
	}
	return c.WriteJson(map[string]any{
		"data": map[string]uint{
			"44":  0,
			"45":  1,
			"46":  13107265,
			"47":  0,
			"66":  1,
			"316": 0,
		},
		"mt": "SS",
	})
}

func v1RdHandler(c *Connection, _ []byte) error {
	if err := c.WriteJson(rfData(0)); err != nil {
		return nil
	}

	if err := c.WriteJson(rfData(0)); err != nil {
		return nil
	}

	return c.WriteJson(map[string]any{
		"data": map[string]any{
			"44":  0,
			"45":  1,
			"46":  65,
			"47":  0,
			"66":  1,
			"316": 0,
		},
		"mt": "SS",
	})
}

func rfData(n uint8) map[string]any {
	return map[string]any{
		"data": map[string]uint8{
			"42": n,
		},
		"mt": "RF",
	}
}
