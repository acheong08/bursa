package ws

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"nhooyr.io/websocket"
)

type MessageHandler func(*Connection, []byte) error

func newConnection(ctx context.Context, conn *websocket.Conn) Connection {
	ctx, cancel := context.WithCancel(ctx)
	return Connection{
		conn:     conn,
		ctx:      ctx,
		handlers: make(map[string]MessageHandler),
		timeout: time.AfterFunc(30*time.Second, func() {
			cancel()
		}),
	}
}

type Connection struct {
	conn     *websocket.Conn
	handlers map[string]MessageHandler
	ctx      context.Context
	timeout  *time.Timer
}

func (c *Connection) StartReadLoop() error {
	for {
		select {
		case <-c.ctx.Done():
			return nil
		default:
			mt, data, err := c.conn.Read(c.ctx)
			if err != nil {
				return err
			}
			if mt != websocket.MessageText {
				log.Println("Received message of type ", mt.String())
				continue
			}
			if err := c.HandleMessage(data); err != nil {
				return err
			}
		}
	}
}

func (c *Connection) WriteJson(m map[string]any) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return c.conn.Write(c.ctx, websocket.MessageText, data)
}

func (c *Connection) AddHandler(n string, f MessageHandler) {
	c.handlers[n] = f
}

func (c *Connection) HandleMessage(m []byte) error {
	var messages []map[string]any
	if err := json.Unmarshal(m, &messages); err != nil {
		log.Println(string(m))
		return err
	}
	for _, message := range messages {
		mt, ok := message["mt"].(string)
		if !ok {
			return errors.New("no message type found")
		}
		if _, ok := message["data"]; !ok {
			return errors.New("no data found")
		}
		mb, _ := json.Marshal(message["data"])
		if f, ok := c.handlers[mt]; ok {
			if err := f(c, mb); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Connection) Subscribe(ids []uint) error {
	return c.WriteJson(map[string]any{
		"data": map[string]any{
			"1":   ids,
			"20":  255,
			"21":  []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			"52":  0,
			"60":  []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			"66":  1,
			"67":  1,
			"69":  []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			"229": len(ids),
		},
		"mt": "RS",
	})
}
