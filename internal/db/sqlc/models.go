// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package db

import (
	"time"
)

type Lichess struct {
	ID        int32     `json:"id"`
	LichessID string    `json:"lichess_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type TgbotUser struct {
	ID       int64 `json:"id"`
	Isactive bool  `json:"isactive"`
}
