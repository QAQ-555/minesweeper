package service

import (
	"context"
	"encoding/json"
	"minesweeper/app/model"

	"minesweeper/app/logic"
)

func ProcessGameCtrlPayload(ctx context.Context, payload json.RawMessage, c *model.Client) error {
	var ctrl model.GameCtrlpayload
	if err := json.Unmarshal(payload, &ctrl); err != nil {
		return err
	}

	if ctrl.Click {
		model.Logger.Infof(ctx, "enter left")
		logic.HandleLeftClick(ctrl.X, ctrl.Y, c)
	} else {
		model.Logger.Infof(ctx, "enter right")
		//logic.HandleRightClick(ctrl.X, ctrl.Y)
	}
	return nil
}
