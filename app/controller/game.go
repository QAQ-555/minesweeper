package controller

import (
	"context"
	"encoding/json"
	"minesweeper/app/logic"
	"minesweeper/app/model"
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
