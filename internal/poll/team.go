package poll

import (
	"context"
	"log/slog"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	db "github.com/swahili-chess/sw-chessbot/internal/db/sqlc"
	lichess "github.com/swahili-chess/sw-chessbot/internal/lichess"
)

type SWbot struct {
	Bot   *tgbotapi.BotAPI
	Store db.Store
	Links *map[string]time.Time
	mu    sync.RWMutex
}

func (sw *SWbot) PollTeam(playersId chan<- []db.InsertLichessDataParams) {

	ticker := time.NewTicker(time.Minute * 5)

	defer ticker.Stop()

	for range ticker.C {

		usernames := lichess.FetchTeamMembers()

		playersId <- lichess.FetchTeamMembers()

		sw.InsertUsernames(usernames)

	}
}

func (sw *SWbot) InsertUsernames(list []db.InsertLichessDataParams) {
	// get current usernames in db
	lichess_ids, err := sw.Store.GetLichessData(context.Background())

	if err != nil {
		slog.Error("Failed to get usernames in DB")
		return
	}

	newPlayers := findNewPlayers(lichess_ids, list)

	for _, player := range newPlayers {

		err := sw.Store.InsertLichessData(context.Background(), player)

		if err != nil {
			slog.Error("Failed to insert user", "player", player)
		}

	}
}

func findNewPlayers(lichess_ids []string, players []db.InsertLichessDataParams) []db.InsertLichessDataParams {
	newPlayers := []db.InsertLichessDataParams{}
	elementSet := make(map[string]bool)

	for _, lichess_id := range lichess_ids {
		elementSet[lichess_id] = true
	}

	for _, dt := range players {
		if _, found := elementSet[dt.LichessID]; !found {
			newPlayers = append(newPlayers, dt)
		} else {
			delete(elementSet, dt.LichessID) // Remove common elements
		}
	}

	return newPlayers
}
