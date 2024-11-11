package poll

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/swahili-chess/sw-chessbot/config"
	lichess "github.com/swahili-chess/sw-chessbot/internal/lichess"
	"github.com/swahili-chess/sw-chessbot/internal/req"
)

type SWbot struct {
	Bot   *tgbotapi.BotAPI
	Links *map[string]time.Time
	mu    sync.RWMutex
}

func (sw *SWbot) PollTeam(members chan<- []lichess.InsertMemberParams) {

	ticker := time.NewTicker(time.Minute * 5)

	defer ticker.Stop()

	for range ticker.C {
		res := lichess.FetchTeamMembers()
		members <- res
		sw.InsertNewMembers(res)

	}
}

func (sw *SWbot) InsertNewMembers(allMembers []lichess.InsertMemberParams) {

	var oldMembers []string
	var errResponse req.ErrorResponse

	statusCode, err := req.GetRequest(fmt.Sprintf("%s/lichess/members", config.Cfg.Url), &oldMembers, &errResponse)

	if statusCode == http.StatusOK && err == nil {
		newMembers := findNewMembers(oldMembers, allMembers)
		for _, player := range newMembers {
			statusCode, err := req.PostOrPutRequest(http.MethodPost,fmt.Sprintf("%s/lichess/members", config.Cfg.Url), player, &errResponse)
			if statusCode == http.StatusInternalServerError {
				slog.Error("Failed to insert member", "error", errResponse.Error)
			} else if err != nil {
				slog.Error("Failed to insert member", "error", err)
			}
		}
	} else if statusCode == http.StatusInternalServerError {
		slog.Error("Failed to get members", "error", errResponse.Error)

	} else {
		slog.Error("Failed to get members", "error", err)
	}

}

func findNewMembers(oldMembers []string, allMembers []lichess.InsertMemberParams) []lichess.InsertMemberParams {
	newMembers := []lichess.InsertMemberParams{}
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
