package model

import (
	"encoding/json"
	"net/http"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gorilla/websocket"
)

type WebMessage struct {
	Type    byte            `json:"type"`
	ID      string          `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

type Client struct {
	ID           string
	Conn         *websocket.Conn
	MapServer    [][]bool
	MapClient    [][]byte
	Map_size_x   uint
	Map_size_y   uint
	Map_mine_num uint
}

type GameCtrlpayload struct {
	X     uint `json:"x"`
	Y     uint `json:"y"`     //click coord on MapClient
	Click bool `json:"click"` //click left or right
}

const (
	TypeOrigin byte = 0x00
	TypeCtrl   byte = 0x0F
)

const (
	WAIT_REPLY_TIME = 60
)

var (
	Logger     = g.Log()
	WsUpGrader = websocket.Upgrader{
		// CheckOrigin allows any origin in development
		// In production, implement proper origin checking for security
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		// Error handler for upgrade failures
		Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
			// Implement error handling logic here
		},
	}
)
