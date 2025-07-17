package service

import (
	"context"
	"encoding/json"
	"fmt"
	"minesweeper/app/model"

	"minesweeper/app/logic"

	"github.com/google/uuid"
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
		if c.MapServer[ctrl.Y][ctrl.X] == model.MineCell {
			fmt.Println("game end")
			// 可以在这里结束游戏逻辑
			return nil
		}
		result := make([]model.ClickResultpayload, 0)
		logic.HandleLeftClick(ctrl.X, ctrl.Y, c, &result)
		if len(result) != 0 {
			msgChainId := uuid.NewString()
			for i := range result {
				isEnd := (i == len(result)-1)
				result[i].IsEnd = isEnd
				result[i].MsgId = msgChainId
				data, err := PackWebMessageJson(ctx, model.TypeResult, result[i], "")
				if err != nil {
					model.Logger.Panicf(ctx, "unkonw fail!!")
				}
				err = c.Conn.WriteMessage(1, data)
				if err != nil {
					model.Logger.Errorf(ctx, "send massage fail")
				}
			}
		}
	} else {
		model.Logger.Infof(ctx, "enter right")
		//logic.HandleRightClick(ctrl.X, ctrl.Y)
	}
	return nil
}
