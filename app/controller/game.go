package controller

import (
	"context"
	"encoding/json"
	"minesweeper/app/logic"
	"minesweeper/app/model"
	"minesweeper/app/service"

	"github.com/gorilla/websocket"
)

func ProcessGameCtrlPayload(ctx context.Context, payload json.RawMessage, c *model.Client) error {
	var ctrl model.GameCtrlpayload
	if err := json.Unmarshal(payload, &ctrl); err != nil {
		return err
	}
	//when first click make map to make sure not first click game fail
	if c.MapServer == nil {
		board, err := logic.CreateRealMap(c.Map_size_x, c.Map_size_y, c.Map_mine_num, ctrl.X, ctrl.Y)
		model.Logger.Debugf(ctx, "create map %d %d %d", c.Map_size_x, c.Map_size_y, c.Map_mine_num)
		if err != nil {
			model.Logger.Errorf(ctx, "failed to create map: %v", err)
			return err
		}
		c.MapServer = board
		logic.SaveBoardWithCoords(ctx, board)
	}

	if ctrl.Click {
		model.Logger.Infof(ctx, "enter left")
		logic.HandleLeftClick(ctx, ctrl.X, ctrl.Y, c)
	} else {
		model.Logger.Infof(ctx, "enter right")
		logic.HandleRightClick(ctx, ctrl.X, ctrl.Y, c)
	}
	return nil
}

func ProcessGameRemakePayload(ctx context.Context, payload json.RawMessage, c *model.Client) error {
	var remake model.GameOptionPayload
	if err := json.Unmarshal(payload, &remake); err != nil {
		return err
	}
	c.MapClient = make([][]byte, remake.Y)
	for i := range c.MapClient {
		c.MapClient[i] = make([]byte, remake.X)
		for j := range c.MapClient[i] {
			c.MapClient[i][j] = model.Unknown
		}
	}
	c.Map_size_x = remake.X
	c.Map_size_y = remake.Y
	c.Map_mine_num = remake.MineNUM
	c.MapServer = nil

	mes, err := service.PackWebMessageJson(ctx, model.TypeOrigin, c, c.ID)
	if err != nil {
		model.Logger.Errorf(ctx, "failed to pack web message: %v", err)
		return err
	}

	err = c.Conn.WriteMessage(websocket.TextMessage, mes)
	if err != nil {
		model.Logger.Errorf(ctx, "failed to write message: %v", err)
		return err
	}
	model.Logger.Infof(ctx, "remake map %d %d %d", c.Map_size_x, c.Map_size_y, c.Map_mine_num)
	return nil
}
