package ws

import (
	"bursa-alert/lib/utils"
	"context"
	"crypto/tls"
	"log"
	"net/http"

	"nhooyr.io/websocket"
)

var headers http.Header = map[string][]string{
	"Accept":                   {"*/*"},
	"Accept-Encoding":          {"gzip, deflate, br"},
	"Accept-Language":          {"en-US,en;q=0.5"},
	"Cache-Control":            {"no-cache"},
	"Connection":               {"keep-alive, Upgrade"},
	"DNT":                      {"1"},
	"Host":                     {"mbbpfs1600.cyberstock.com.my"},
	"Origin":                   {"https://ost.maybank2u.com.my"},
	"Pragma":                   {"no-cache"},
	"Sec-Fetch-Dest":           {"websocket"},
	"Sec-Fetch-Mode":           {"websocket"},
	"Sec-Fetch-Site":           {"cross-site"},
	"Sec-WebSocket-Extensions": {"permessage-deflate"},
	// "Sec-WebSocket-Key":        {"LdSCw1q/Ggfve5nnbcOqRg=="},
	// "Sec-WebSocket-Version":    {"13"},
	"Upgrade":    {"websocket"},
	"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; rv:114.0) Gecko/20100101 Firefox/114.0"},
}

type connectionOptions struct {
	// v1           bool
	initHandlers map[string]MessageHandler
}

func defaultOptions() connectionOptions {
	return connectionOptions{initHandlers: make(map[string]MessageHandler)}
}

type OptionModifier func(*connectionOptions)

//
// var WithUseV3 OptionModifier = func(co *connectionOptions) {
// 	co.v1 = true
// }

func WithMessageHandler(mt string, mh MessageHandler) OptionModifier {
	return func(co *connectionOptions) {
		co.initHandlers[mt] = mh
	}
}

func NewConnection(ctx context.Context, subscriptions []uint, options ...OptionModifier) (Connection, error) {
	opts := defaultOptions()
	for _, opt := range options {
		opt(&opts)
	}

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				CipherSuites: []uint16{
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				},
			},
		},
	}

	c, _, err := websocket.Dial(ctx, "wss://mbbpfs2600.cyberstock.com.my/", &websocket.DialOptions{
		HTTPClient: &client,
		HTTPHeader: headers,
	})
	if err != nil {
		return Connection{}, err
	}
	conn := newConnection(ctx, c)
	if err := conn.WriteJson(map[string]any{
		"data": map[string]any{
			"14":  "this is hash value",
			"37":  4325,
			"64":  196608,
			"65":  2022,
			"271": "3b456bb11a511fcd3fa0b6ccf05faf3f5a9809cffd6d5230b09b13aa1c171db3",
			"16":  utils.RandInt(12),
		},
		"mt": "LG",
	}); err != nil {
		return Connection{}, err
	}
	conn.AddHandler("MS", pingHandler)
	// if opts.v1 {
	conn.AddHandler("RD", v3RdHandler)
	// } else {
	// 	conn.AddHandler("RD", v1RdHandler)
	// }
	for mt, handler := range opts.initHandlers {
		conn.AddHandler(mt, handler)
	}

	subscribed := false
	conn.AddHandler("SU", func(conn *Connection, _ []byte) error {
		subscribed = true
		return conn.Subscribe(subscriptions)
	})
	// Wait for mt:SU (ready for subscribe)
	for {
		mt, data, err := conn.conn.Read(ctx)
		if err != nil {
			return Connection{}, err
		}
		if mt != websocket.MessageText {
			continue
		}
		if err = conn.HandleMessage(data); err != nil {
			return Connection{}, err
		}
		if subscribed {
			log.Println("Subscription complete")
			break
		}

	}

	return conn, nil
}
