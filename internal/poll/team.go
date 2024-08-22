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

func (sw *SWbot) PollTeam(members chan<- []db.InsertMemberParams) {

	ticker := time.NewTicker(time.Minute * 5)

	defer ticker.Stop()

	for range ticker.C {
		res := lichess.FetchTeamMembers()
		members <- res
		sw.InsertNewMembers(res)

	}
}

func (sw *SWbot) InsertNewMembers(allMembers []db.InsertMemberParams) {

	oldMembers, err := sw.Store.GetLichessMembers(context.Background())
	if err != nil {
		slog.Error("Failed to get usernames in DB")
		return
	}

	newMembers := findNewMembers(oldMembers, allMembers)
	for _, player := range newMembers {
		err := sw.Store.InsertMember(context.Background(), player)
		if err != nil {
			slog.Error("Failed to insert user", "player", player)
		}

	}
}

func findNewMembers(oldMembers []string, allMembers []db.InsertMemberParams) []db.InsertMemberParams {
	newMembers := []db.InsertMemberParams{}
	oldMembersSet := make(map[string]bool)

	for _, member := range oldMembers {
		oldMembersSet[member] = true
	}

	for _, member := range allMembers {
		if _, found := oldMembersSet[member.LichessID]; !found {
			newMembers = append(newMembers, member)
		} else {
			delete(oldMembersSet, member.LichessID) 
		}
	}

	return newMembers
}
