package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"minesweeper/app/controller"
	"minesweeper/app/logic"
	"minesweeper/app/model"
	"minesweeper/app/service"
	msgpack "minesweeper/test"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func Handler(r *ghttp.Request) {
	ws, err := model.WsUpGrader.Upgrade(r.Response.Writer, r.Request, nil)
	if err != nil {
		r.Response.Write(err.Error())
		return
	}

	// Get request context for logging
	var ctx = r.Context()
	userID := uuid.New().String() // Generate a new UUID for the user
	ok, option := waitForOption(ctx, ws)

	if !ok {
		log.Println("⏳ 超时或失败，未获取 username")
		ws.Close() // 关闭连接，释放资源
		return
	}

	//create a real minesweeper map just have mine or not
	var x, y, mineNum uint = option.X, option.Y, option.MineNUM

	// Initialize user map with zeros
	// This map will be used to track the user's flags and revealed cells
	userMap := make([][]byte, y)
	for i := range userMap {
		userMap[i] = make([]byte, x)
		for j := range userMap[i] {
			userMap[i][j] = model.Unknown
		}
	}

	client := &model.Client{
		ID:           userID,
		Conn:         ws,
		MapClient:    userMap,
		MapServer:    nil,
		Map_size_x:   x,
		Map_size_y:   y,
		Map_mine_num: mineNum,
	}

	mes, err := service.PackWebMessageJson(ctx, model.TypeOrigin, client, userID)
	if err != nil {
		model.Logger.Errorf(ctx, "failed to pack web message: %v", err)
		return
	}

	client.Conn.WriteMessage(websocket.TextMessage, mes)

	// Message handling loop
	go handleClientMessages(ctx, client)

}

func handleClientMessages(ctx context.Context, client *model.Client) {
	defer func() {
		model.Logger.Info(ctx, "free resource for client: %s", client.ID)
		client.Conn.Close()
		model.Logger.Info(ctx, "websocket connection closed")
		logic.DeleteBoardFile(ctx)
	}()
	// Message handling loop
	for {
		// Read incoming WebSocket message
		_, msg, err := client.Conn.ReadMessage()
		if err != nil {
			break // Connection closed or error occurred
		}

		gameMsgtype, id, payload, err := service.UnpackWebMessage(msg)
		if err != nil {
			log.Printf("Failed to parse JSON from %s: %v", client.ID, err)
			continue
		}
		model.Logger.Infof(ctx, "received message type: %d, id : %s ,payload: %s", gameMsgtype, id, payload)

		switch gameMsgtype {

		case model.TypeCtrl:
			model.Logger.Infof(ctx, "enter ctrl")
			err = controller.ProcessGameCtrlPayload(ctx, payload, client)
		case model.TypeOrigin:
			model.Logger.Infof(ctx, "enter remake")
			err = controller.ProcessGameRemakePayload(ctx, payload, client)
		default:
			model.Logger.Errorf(ctx, "unknown message type: %d", gameMsgtype)
			continue
		}

		if err != nil {
			model.Logger.Errorf(ctx, "Failed to parse payload from %s: %v", client.ID, err)
			continue
		}
	}
}

// 等待客户端注册用户名
func waitForOption(ctx context.Context, ws *websocket.Conn) (bool, *model.GameOptionPayload) {

	ws.SetReadDeadline(time.Now().Add(model.WAIT_REPLY_TIME * time.Second))
	msgCh := make(chan []byte)
	timeoutCh := make(chan bool)
	closeCh := make(chan bool)

	// 读取消息的 goroutine
	go func() {
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				timeoutCh <- true
				<-closeCh
				return
			}

			msgCh <- msg
			val := <-closeCh
			if val {
				return
			}
		}
	}()

	defer func() {
		ws.SetReadDeadline(time.Time{})
		close(closeCh)
	}()

	for {
		select {
		case <-timeoutCh:
			closeCh <- true
			return false, nil

		case msg := <-msgCh:
			readNext, option, err := handleRegisterMessage(ctx, msg)
			closeCh <- readNext

			if readNext {
				return true, option
			} else {
				if err != nil {
					// 可按需保留业务错误处理
				}
			}
		}
	}
}

// 处理注册消息
func handleRegisterMessage(ctx context.Context, msg []byte) (bool, *model.GameOptionPayload, error) {

	msgbuf := &msgpack.MsgPack{}

	err := proto.Unmarshal(msg, msgbuf)
	if err != nil {
		model.Logger.Errorf(ctx, "failed to unmarshal protobuf message: %v", err)
		msgType, _, payload, err := service.UnpackWebMessage(msg)

		if err != nil || msgType != model.TypeOrigin {
			return false, nil, fmt.Errorf("failed to parse message: %w", err)
		}
		var gameoption model.GameOptionPayload
		if err := json.Unmarshal(payload, &gameoption); err != nil {
			return false, nil, err
		}
		return true, &gameoption, err
	} else {
		payload := msgbuf.GetGameInitPayload()
		if payload == nil {
			return false, nil, fmt.Errorf("invalid game init payload")
		}
		return true, &model.GameOptionPayload{
			X:       uint(payload.GetX()),
			Y:       uint(payload.GetY()),
			MineNUM: uint(payload.GetMineNum()),
		}, nil
	}
}
