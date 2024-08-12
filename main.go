package main

import (
	"bursa-alert/internal/database"
	"bursa-alert/lib"
	"bursa-alert/lib/alerts"
	"bursa-alert/lib/global"
	"bursa-alert/lib/models"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

//go:embed frontend/*
var frontend embed.FS

var (
	notificationsCache = make(map[uint]models.StockEntry)
	stockMetadata      = make(map[uint]models.StockMetadata)
)

func main() {
	wsMap := make(map[uint]*websocket.Conn)
	wsIndex := uint(0)

	alertList := database.LoadAlerts()
	e := echo.New()

	subFs, _ := fs.Sub(frontend, "frontend")
	e.GET("/*", echo.WrapHandler(http.FileServerFS(subFs)))
	// Alert GET/POST/DELETE
	alertGroup := e.Group("/alerts")
	alertGroup.GET("/", func(c echo.Context) error {
		return c.JSON(200, alertList)
	})
	alertGroup.PUT("/", func(c echo.Context) error {
		alert := alerts.Alert{}
		if err := c.Bind(&alert); err != nil {
			return c.String(400, "Failed to bind alert")
		}
		if err := alert.Validate(); err != nil {
			return c.String(400, fmt.Sprintf("Failed to parse alert: %s", err))
		}
		alertList = append(alertList, alert)
		return c.JSON(200, alertList)
	})
	alertGroup.POST("/", func(c echo.Context) error {
		id, err := func() (int, error) {
			id := c.QueryParam("id")
			return strconv.Atoi(id)
		}()
		if err != nil {
			return c.String(400, "Missing index")
		}
		if id >= len(alertList) {
			return c.String(400, "Index out of range")
		}
		alert := alerts.Alert{}
		if err := c.Bind(&alert); err != nil {
			return c.String(400, "Failed to bind alert")
		}
		if err := alert.Validate(); err != nil {
			return c.String(400, fmt.Sprintf("Failed to parse alert: %s", err))
		}
		alertList[id] = alert
		return c.JSON(200, alertList)
	})
	alertGroup.DELETE("/", func(c echo.Context) error {
		index, err := strconv.Atoi(c.QueryParam("id"))
		if err != nil {
			return c.String(400, "Failed to parse index")
		}

		if len(alertList) < index {
			return c.String(400, "Index out of range")
		}
		alertList = append(alertList[:index], alertList[index+1:]...)
		return c.String(404, "Alert not found")
	})
	// Websocket for alerts
	e.GET("/ws", func(c echo.Context) error {
		ws, err := websocket.Accept(c.Response().Writer, c.Request(), nil)
		if err != nil {
			return c.String(400, "Failed to upgrade connection")
		}
		defer ws.Close(websocket.StatusNormalClosure, "")
		ctx, cancel := context.WithCancel(c.Request().Context())
		defer cancel()
		timer := time.AfterFunc(10*time.Second, func() {
			cancel()
		})
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					err := wsjson.Write(ctx, ws, map[string]string{"action": "ping"})
					if err != nil {
						return
					}
					time.Sleep(5 * time.Second)
				}
			}
		}()

		wsMap[wsIndex] = ws
		wsIndex++
		defer delete(wsMap, wsIndex)

		// Send notification cache
		for _, entry := range notificationsCache {
			_ = wsjson.Write(ctx, ws, map[string]any{
				"id":     entry.GetIndex(),
				"data":   entry.ToMap(),
				"ticker": stockMetadata[entry.GetIndex()].Ticker,
				"name":   stockMetadata[entry.GetIndex()].Name,
			})
		}

		for {
			// Wait for pong
			select {
			case <-ctx.Done():
				return nil
			default:
				var msg map[string]string
				err := wsjson.Read(ctx, ws, &msg)
				if err != nil {
					return nil
				}
				if msg, ok := msg["action"]; ok && msg == "pong" {
					timer.Reset(10 * time.Second)
				}
			}
		}
	})
	go stockings(alertList, wsMap)

	go func() {
		if err := e.Start("127.0.0.1:1970"); err != nil {
			panic(err)
		}
	}()
	// Wait for interrupt to exit gracefully with defered stuff
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	database.SaveAlerts(alertList)
	println("Exiting")
}

func stockings(alertList []alerts.Alert, wsMap map[uint]*websocket.Conn) {
	// Create initial connection to fetch metadata
	ctx, cancel := context.WithCancel(context.Background())
	if err := lib.GetStockMetadata(stockMetadata); err != nil {
		panic(err)
	}
	cancel()
	stockCh := make(chan models.StockEntry, 100)
	errCh := make(chan error)
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	ids := make([]uint, len(stockMetadata))
	i := 0
	for key := range stockMetadata {
		ids[i] = key
		i++
	}
	for i := 0; i <= len(ids); i += 50 {
		go func(i int) {
			for {
				var err error
				if (i + 100) > len(ids) {
					err = lib.GetDataStream(ctx, ids[i:], stockCh)
				} else {
					err = lib.GetDataStream(ctx, ids[i:i+50], stockCh)
				}
				if err != nil {
					errCh <- err
					continue
				}
			}
		}(i)
	}
	// err = lib.GetDataStream(ctx, ids[0:50], stockCh, errCh)
	// if err != nil {
	// 	panic(err)
	// }
	for {
		select {
		case stock := <-stockCh:
			stock = global.Entries.Push(stock)
			for _, alert := range alertList {
				if alert.Eval(stock.GetIndex()) {
					notificationsCache[stock.GetIndex()] = stock
					for _, ws := range wsMap {
						_ = wsjson.Write(ctx, ws, map[string]any{
							"id":     stock.GetIndex(),
							"data":   stock.ToMap(),
							"ticker": stockMetadata[stock.GetIndex()].Ticker,
							"name":   stockMetadata[stock.GetIndex()].Name,
							"alert":  alert.Label,
						})
					}
				}
			}
		case err := <-errCh:
			fmt.Println(err)
			break
		}
	}
}
