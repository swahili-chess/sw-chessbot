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

func (sw *SWbot) PollTeam(playersId chan<- []db.InsertMemberParams) {

	ticker := time.NewTicker(time.Minute * 5)

	defer ticker.Stop()

	for range ticker.C {

		usernames := lichess.FetchTeamMembers()

		playersId <- lichess.FetchTeamMembers()

		sw.InsertNewMembers(usernames)

	}
}

func (sw *SWbot) InsertNewMembers(list []db.InsertMemberParams) {
	// get current usernames in db
	lichess_ids, err := sw.Store.GetLichessMembers(context.Background())

	if err != nil {
		slog.Error("Failed to get usernames in DB")
		return
	}

	newMembers := findNewMembers(lichess_ids, list)

	for _, player := range newMembers {

		err := sw.Store.InsertMember(context.Background(), player)

		if err != nil {
			slog.Error("Failed to insert user", "player", player)
		}

	}
}

func findNewMembers(lichess_ids []string, players []db.InsertMemberParams) []db.InsertMemberParams {
	newMembers := []db.InsertMemberParams{}
	elementSet := make(map[string]bool)

	for _, lichess_id := range lichess_ids {
		elementSet[lichess_id] = true
	}

	for _, dt := range players {
		if _, found := elementSet[dt.LichessID]; !found {
			newMembers = append(newMembers, dt)
		} else {
			delete(elementSet, dt.LichessID) // Remove common elements
		}
	}

	return newMembers
}
