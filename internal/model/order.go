package model

import (
	"encoding/json"
	"time"
)

const InProgress = "in progress"

type OrderMsg struct {
	ActionID string  `json:"action_id"`
	ItemIDs  []int64 `json:"item_ids"`
}

type OrderFromQueue struct {
	ActionID string    `json:"action_id"`
	Time     time.Time `json:"time"`
}

func (o *OrderMsg) ToJson() (string, error) {

	// marshal to json
	b, err := json.Marshal(o)

	if err != nil {
		return "", err
	}

	return string(b), nil
}
