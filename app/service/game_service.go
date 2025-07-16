package service

import (
	"context"
	"encoding/json"
	"minesweeper/app/model"

	"minesweeper/app/logic"
)

func ProcessGameCtrlPayload(ctx context.Context, payload json.RawMessage) error {
	var ctrl model.GameCtrlpayload
	if err := json.Unmarshal(payload, &ctrl); err != nil {
		return err
	}
	if ctrl.Click {
		logic.HandleLeftClick(ctrl.X, ctrl.Y)
	} else {
		logic.HandleRightClick(ctrl.X, ctrl.Y)
	}
	return nil
}
