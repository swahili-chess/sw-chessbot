package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	withGameIds      = "&withGameIds=true"
	urlStatus        = "https://lichess.org/api/users/status?ids="
	base_url         = "https://lichess.org/"
	minLinkStayInMap = 1 * time.Hour
	cleanUpTime      = 30 * time.Minute
)

type PlayerStatus struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Title     string `json:"title,omitempty"`
	Online    bool   `json:"online"`
	Playing   bool   `json:"playing"`
	Streaming bool   `json:"streaming"`
	Patron    bool   `json:"patron"`
	PlayingId string `json:"playingId"`
}

func (sw *SWbot) sendMessagesToIds(linkId string) {
	gameLink := base_url + linkId

	ids, _ := sw.models.Users.GetActiveUsers()

	for _, id := range ids {
		msg := tgbotapi.NewMessage(id.Id, gameLink)

		sw.bot.Send(msg)
	}
}

func (sw *SWbot) fetchPlayersStatus(url string, links *map[string]time.Time) {
	var playerStatuses []PlayerStatus
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error while fetching status")
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&playerStatuses)

	if err != nil {
		log.Println("Error decoding the json body", err)
	}

	for _, playerStatus := range playerStatuses {
		if len(playerStatus.PlayingId) != 0 {

			sw.mu.RLock()
			_, idExists := (*links)[playerStatus.PlayingId]
			sw.mu.RUnlock()

			if !idExists {
				sw.mu.Lock()
				(*links)[playerStatus.PlayingId] = time.Now()
				sw.mu.Unlock()

				sw.sendMessagesToIds(playerStatus.PlayingId)
			}

		}
	}

}

func prepareFetchStatusUrl(playersIds []string) string {

	joinedPlayerIds := strings.Join(playersIds, ",")

	fetchStatusUrl := urlStatus + joinedPlayerIds + withGameIds

	return fetchStatusUrl

}

func (sw *SWbot) cleanUpMap(links *map[string]time.Time) {

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


func (sw *SWbot) sendMaintananceMsg(msg string) {

	ids, _ := sw.models.Users.GetActiveUsers()

	for _, id := range ids {
		msg := tgbotapi.NewMessage(id.Id,msg)

		sw.bot.Send(msg)
	}
}
