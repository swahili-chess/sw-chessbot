package poll

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/swahili-chess/sw-chessbot/internal/lichess"
	"github.com/swahili-chess/sw-chessbot/internal/req"
)

const (
	withGameIds      = "&withGameIds=true"
	urlStatus        = "https://lichess.org/api/users/status?ids="
	base_url         = "https://lichess.org/"
	minLinkStayInMap = 1 * time.Hour
	cleanUpTime      = 30 * time.Minute
	Master_ID        = 731217828
)

type Member struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Title     string `json:"title,omitempty"`
	Online    bool   `json:"online"`
	Playing   bool   `json:"playing"`
	Streaming bool   `json:"streaming"`
	Patron    bool   `json:"patron"`
	PlayingId string `json:"playingId"`
}

func (sw *SWbot) PollMember(membersIdsChan <-chan []lichess.InsertMemberParams, membersIds *[]lichess.InsertMemberParams) {

	ticker := time.NewTicker(time.Second * 6)
	defer ticker.Stop()

	go sw.cleanUpMap(sw.Links)

	for range ticker.C {
		select {

		case memberIdsLists := <-membersIdsChan:
			if len(memberIdsLists) != 0 {
				*membersIds = memberIdsLists
			}

		default:
			url := prep_url(*membersIds, urlStatus)
			if url != "" {
				sw.fetchMemberDetails(url, sw.Links)
			}
		}

	}
}

// Fetch the status of the players whether they are playing or not
func (sw *SWbot) fetchMemberDetails(url string, links *map[string]time.Time) {
	var members []Member

	var client = &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Get(url)
	if err != nil {
		slog.Error("Error while fetching Member status details", "err", err)
		return
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&members)
	if err != nil {
		slog.Error("Error decoding the json body", "err", err)
		return
	}

	for _, member := range members {

		// Sometimes the playingId is empty string
		if len(member.PlayingId) != 0 {

			sw.mu.RLock()
			_, idExists := (*links)[member.PlayingId]
			sw.mu.RUnlock()

			if !idExists {
				sw.mu.Lock()
				(*links)[member.PlayingId] = time.Now()
				sw.mu.Unlock()

				sw.SendMsgToTelegramIds(member.PlayingId)
			}

		}
	}

}

// Prepare the url to fetch the status of the members
func prep_url(members []lichess.InsertMemberParams, urlStatus string) string {

	if len(members) == 0 {
		return ""
	}
	membersIds := []string{}
	for _, member := range members {
		membersIds = append(membersIds, member.LichessID)
	}

	var urlBuilder strings.Builder
	urlBuilder.WriteString(urlStatus)
	urlBuilder.WriteString(strings.Join(membersIds, ","))
	urlBuilder.WriteString(withGameIds)

	return urlBuilder.String()

}

// Delete links that have stayed in the map for more than 1 hour
func (sw *SWbot) cleanUpMap(links *map[string]time.Time) {

	// Run the clean up every 30 minutes
	ticker := time.NewTicker(cleanUpTime)

	defer ticker.Stop()

	for range ticker.C {
		for lichessId, timeAtStart := range *links {
			if time.Since(timeAtStart) > minLinkStayInMap {
				sw.mu.Lock()
				delete(*links, lichessId)
				sw.mu.Unlock()

			}
		}
	}
}

func (sw *SWbot) SendMsgToTelegramIds(linkId string) {

	var ids []int64
	var errResponse req.ErrorResponse

	statusCode, err := req.GetRequest("https://api.swahilichess.com/telegram/bot/users/active", &ids, &errResponse)
	if statusCode != http.StatusInternalServerError {
		slog.Error("failed to get telegram bot users", "err", errResponse.Error)
		
	} else if statusCode != http.StatusOK || err != nil {
		slog.Error("failed to get telegram bot users", "err", err, "statusCode", statusCode)
	}
	for _, id := range ids {
		msg := tgbotapi.NewMessage(id, fmt.Sprintf("%s%s", base_url, linkId))

		sw.Bot.Send(msg)
	}
}

// Send a message to all active users when the bot is going for maintanance
func (sw *SWbot) SendMaintananceMsg(msg string) {

	var ids []int64
	var errResponse req.ErrorResponse

	statusCode, err := req.GetRequest("https://api.swahilichess.com/telegram/bot/users/active", &ids, &errResponse)
	if statusCode != http.StatusInternalServerError {
		slog.Error("failed to get telegram bot users", "err", errResponse.Error)

	} else if statusCode != http.StatusOK || err != nil {
		slog.Error("failed to get telegram bot users", "err", err, "statusCode", statusCode)

	}
	for _, id := range ids {
		if id != Master_ID {
			msg := tgbotapi.NewMessage(id, msg)
			sw.Bot.Send(msg)
		}

	}
}
