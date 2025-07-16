package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"minesweeper/app/logic"
	"minesweeper/app/model"
	"minesweeper/app/service"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func Handler(r *ghttp.Request) {
	ws, err := model.WsUpGrader.Upgrade(r.Response.Writer, r.Request, nil)
	if err != nil {
		r.Response.Write(err.Error())
		return
	}
	userID := uuid.New().String() // Generate a new UUID for the user
	ok, option := waitForOption(ws)

	if !ok {
		log.Println("⏳ 超时或失败，未获取 username")
		ws.Close() // 关闭连接，释放资源
		return
	}

	// Get request context for logging
	var ctx = r.Context()
	//create a real minesweeper map just have mine or not
	var x, y, mineNum uint = option.X, option.Y, option.MineNUM

	board, err := logic.CreateRealMap(x, y, mineNum)
	if err != nil {
		model.Logger.Errorf(ctx, "failed to create map: %v", err)
		return
	}

	// Initialize user map with zeros
	// This map will be used to track the user's flags and revealed cells
	userMap := make([][]byte, x)
	for i := range userMap {
		userMap[i] = make([]byte, y)
	}

	client := &model.Client{
		ID:        userID,
		Conn:      ws,
		MapServer: board,
		MapClient: &userMap,
	}

	mes, err := service.PackWebMessageJson(ctx, model.TypeOrigin, client, userID)
	if err != nil {
		model.Logger.Errorf(ctx, "failed to pack web message: %v", err)
		return
	}

	client.Conn.WriteMessage(websocket.TextMessage, mes)

	// Message handling loop
	go handleClientMessages(ctx, client)
	// Log connection closure
	model.Logger.Info(ctx, "websocket connection closed")

}

func handleClientMessages(ctx context.Context, client *model.Client) {
	defer func() {
		model.Logger.Info(ctx, "free resource for client: %s", client.ID)
		client.Conn.Close()
	}()
	// Message handling loop
	for {
		// Read incoming WebSocket message
		wsMsgType, msg, err := client.Conn.ReadMessage()
		if err != nil {
			break // Connection closed or error occurred
		}
		gameMsgtype, id, payload, err := service.UnpackWebMessage(msg)
		if err != nil {
			log.Printf("Failed to parse JSON from %s: %v", client.ID, err)
			continue
		}
		model.Logger.Infof(ctx, "received message type: %d, id : %s ,payload: \n %s", gameMsgtype, id, payload)

		switch gameMsgtype {
		case model.TypeCtrl:
			err = service.ProcessGameCtrlPayload(ctx, payload)
		default:
			model.Logger.Errorf(ctx, "unknown message type: %d", gameMsgtype)
			continue
		}

		if err != nil {
			model.Logger.Errorf(ctx, "Failed to parse payload from %s: %v", client.ID, err)
			continue
		}

		// Log received message
		model.Logger.Infof(ctx, "received message: %s", msg)
		// Echo the message back to client
		if err = client.Conn.WriteMessage(wsMsgType, msg); err != nil {
			break // Error writing message
		}
	}
}

// 等待客户端注册用户名
func waitForOption(ws *websocket.Conn) (bool, *model.GameOptionPayload) {

	ws.SetReadDeadline(time.Now().Add(model.WAIT_REPLY_TIME * time.Second))
	msgCh := make(chan []byte)
	timeoutCh := make(chan bool)
	closeCh := make(chan bool)

	// 读取消息的 goroutine
	go func() {
		log.Println("[goroutine] 启动")
		defer log.Println("[goroutine] 退出")

		for {
			// log.Println("[goroutine] 开始 ReadMessage")
			_, msg, err := ws.ReadMessage()
			if err != nil {
				// log.Println("[goroutine] ReadMessage 出错:", err)

				// log.Println("[goroutine] 尝试写入 timeoutCh")
				timeoutCh <- true
				// log.Println("[goroutine] 写入 timeoutCh 成功")
				// log.Println("[goroutine] 等待从 closeCh 读取以完成同步")
				<-closeCh
				// log.Println("[goroutine] 收到 closeCh:", val)
				return
			}

			// log.Println("[goroutine] 读到消息:", string(msg))

			// log.Println("[goroutine] 尝试写入 msgCh")
			msgCh <- msg
			// log.Println("[goroutine] 写入 msgCh 成功")

			// log.Println("[goroutine] 等待从 closeCh 读取")
			val := <-closeCh
			// log.Println("[goroutine] 从 closeCh 收到:", val)
			if val {
				// log.Println("[goroutine] 收到 true，退出")
				return
			}
			// log.Println("[goroutine] 收到 false，继续循环")
		}
	}()

	defer func() {
		// log.Println("[WaitForCondition] defer: 关闭 closeCh")
		ws.SetReadDeadline(time.Time{})
		close(closeCh)
	}()

	for {
		// log.Println("[WaitForCondition] 等待 select")
		select {
		case <-timeoutCh:
			// log.Println("[WaitForCondition] 从 timeoutCh 收到信号")
			// log.Println("[WaitForCondition] 尝试写入 closeCh")
			closeCh <- true
			// log.Println("[WaitForCondition] 写入 closeCh 完成")
			return false, nil

		case msg := <-msgCh:
			// log.Println("[WaitForCondition] 从 msgCh 收到:", string(msg))
			readNext, option, err := handleRegisterMessage(msg)
			// log.Println("[WaitForCondition] processMessage 返回:", readNext, username, err)

			// log.Println("[WaitForCondition] 尝试写入 closeCh:", readNext)
			closeCh <- readNext
			// log.Println("[WaitForCondition] 写入 closeCh 完成")

			if readNext {
				// log.Println("[WaitForCondition] 条件满足，返回 true,", username)
				return true, option
			} else {
				// log.Println("[WaitForCondition] 条件未满足，继续等待")
				if err != nil {
					// log.Println("[WaitForCondition] processMessage 错误:", err)
				}
			}
		}
	}
}

// 处理注册消息
func handleRegisterMessage(msg []byte) (bool, *model.GameOptionPayload, error) {
	msgType, _, payload, err := service.UnpackWebMessage(msg)

	if err != nil || msgType != model.TypeOrigin {
		return false, nil, fmt.Errorf("failed to parse message: %w", err)
	}
	var gameoption model.GameOptionPayload
	if err := json.Unmarshal(payload, &gameoption); err != nil {
		return false, nil, err
	}
	fmt.Printf("(%d %d %d)", gameoption.MineNUM, gameoption.X, gameoption.Y)
	return true, &gameoption, err
}
