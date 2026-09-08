package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/service"
)

type cursorPayload struct {
	CreatedAt time.Time `json:"created_at"`
	ID        int64     `json:"id"`
}

func encodeCursor(
	cursor *service.OrderCursor,
) (string, error) {
	payload := cursorPayload{
		CreatedAt: cursor.CreatedAt,
		ID:        cursor.ID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func decodeCursor(
	value string,
) (*service.OrderCursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}

	var payload cursorPayload

	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	if payload.ID <= 0 || payload.CreatedAt.IsZero() {
		return nil, errors.New("invalid cursor payload")
	}

	return &service.OrderCursor{
		CreatedAt: payload.CreatedAt,
		ID:        payload.ID,
	}, nil
}
