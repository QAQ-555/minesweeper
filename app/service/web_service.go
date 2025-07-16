package service

import (
	"context"
	"encoding/json"
	"minesweeper/app/model"

	"github.com/gogf/gf/v2/frame/g"
)

// PackWebMessageJson packs a message into JSON format for WebSocket communication.
// Parameters:
// - ctx: Context for logging and error handling.
// - msgType: Type of the message (e.g., origin, response).
// - payload: The data to be sent in the message.
// - id: Unique identifier for the message or client.
//
// Returns:
// - A byte slice containing the JSON-encoded message.
// - An error if the packing fails.
func PackWebMessageJson(ctx context.Context, msgType byte, payload interface{}, id string) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		// 处理错误
		g.Log().Errorf(ctx, "failed to marshal client: %v", err)
		return nil, err
	}

	rawMsg := json.RawMessage(data)

	mes := model.WebMessage{
		Type: msgType,
		//TimeStamp: time.Now().UnixNano(),
		ID:      id,
		Payload: rawMsg,
	}

	return json.Marshal(mes)
}

func UnpackWebMessage(msg []byte) (byte, string, json.RawMessage, error) {
	var rmsg model.WebMessage

	err := json.Unmarshal(msg, &rmsg)
	if err != nil {
		g.Log().Errorf(context.Background(), "failed to unmarshal message: %v", err)
		return 0, "", nil, err
	}
	return rmsg.Type, rmsg.ID, rmsg.Payload, nil
}
