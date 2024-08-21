package poll

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	db "github.com/swahili-chess/sw-chessbot/internal/db/sqlc"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	withGameIds      = "&withGameIds=true"
	urlStatus        = "https://lichess.org/api/users/status?ids="
	base_url         = "https://lichess.org/"
	minLinkStayInMap = 1 * time.Hour
	cleanUpTime      = 30 * time.Minute
)

// PlayerInfo is the struct that holds the status of the player
type PlayerInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Title     string `json:"title,omitempty"`
	Online    bool   `json:"online"`
	Playing   bool   `json:"playing"`
	Streaming bool   `json:"streaming"`
	Patron    bool   `json:"patron"`
	PlayingId string `json:"playingId"`
}

func (sw *SWbot) Poller(listOfPlayerIdsChan <-chan []db.InsertLichessDataParams, listOfPlayerIds *[]db.InsertLichessDataParams) {

	ticker := time.NewTicker(time.Second * 6)

	defer ticker.Stop()

	go sw.cleanUpMap(sw.Links)

	for range ticker.C {
		select {

		case playerIdsLists := <-listOfPlayerIdsChan:
			if len(playerIdsLists) != 0 {
				*listOfPlayerIds = playerIdsLists
			}

		default:
			url := prepareFetchInfoUrl(*listOfPlayerIds, urlStatus, withGameIds)

			if url != "" {
				sw.fetchPlayersInfo(url, sw.Links)
			}
		}

	}
}

// Fetch the status of the players whether they are playing or not
func (sw *SWbot) fetchPlayersInfo(url string, links *map[string]time.Time) {
	var listOfPlayerInfos []PlayerInfo

	// Create a new client with a timeout
	var client = &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Get(url)
	if err != nil {
		slog.Error("Error while fetching status")
		return
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&listOfPlayerInfos)

	if err != nil {
		slog.Error("Error decoding the json body", "err", err)
		return
	}

	for _, playerInfo := range listOfPlayerInfos {

		// Sometimes the playingId is empty
		if len(playerInfo.PlayingId) != 0 {

			sw.mu.RLock()
			_, idExists := (*links)[playerInfo.PlayingId]
			sw.mu.RUnlock()

			if !idExists {
				sw.mu.Lock()
				(*links)[playerInfo.PlayingId] = time.Now()
				sw.mu.Unlock()

				sw.SendMsgToTelegramIds(playerInfo.PlayingId)
			}

		}
	}

}

// Prepare the url to fetch the status of the players
func prepareFetchInfoUrl(players []db.InsertLichessDataParams, urlStatus, withGameIds string) string {

	if len(players) == 0 {
		return ""
	}
	playersIds := []string{}

	for _, player := range players {
		playersIds = append(playersIds, player.LichessID)
	}

	joinedPlayerIds := strings.Join(playersIds, ",")

	var urlBuilder strings.Builder
	urlBuilder.WriteString(urlStatus)
	urlBuilder.WriteString(joinedPlayerIds)
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
	gameLink := base_url + linkId

	ids, _ := sw.Store.GetActiveTgBotUsers(context.Background())

	for _, id := range ids {
		msg := tgbotapi.NewMessage(id, gameLink)

		sw.Bot.Send(msg)
	}
}

// Send a message to all active users when the bot is down for maintanance
func (sw *SWbot) SendMaintananceMsg(msg string) {

	ids, _ := sw.Store.GetActiveTgBotUsers(context.Background())

	for _, id := range ids {
		msg := tgbotapi.NewMessage(id, msg)

		sw.Bot.Send(msg)
	}
}
