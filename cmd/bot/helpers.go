package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const withGameIds = "&withGameIds=true"
const urlStatus = "https://lichess.org/api/users/status?ids="
const base_url = "https://lichess.org/"
const minLinkStayInMap = 1 * time.Hour

type UserStatus struct {
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

func (sw *SWbot) fetchStatus(url string, links *map[string]time.Time) {
	var userStatuses []UserStatus
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error while fetching status")
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&userStatuses)

	if err != nil {
		log.Println("Error decoding the json body", err)
	}

	for _, user := range userStatuses {
		if len(user.PlayingId) != 0 {

			sw.mu.RLock()
			_, idExists := (*links)[user.PlayingId]
			sw.mu.RUnlock()

			if !idExists {
				sw.mu.Lock()
				(*links)[user.PlayingId] = time.Now()
				sw.mu.Unlock()

				sw.sendMessagesToIds(user.PlayingId)
			}

		}
	}

}

func prepareUrl(userIds []string) string {

	idsJoined := strings.Join(userIds, ",")

	fetchUrl := urlStatus + idsJoined + withGameIds

	return fetchUrl

}

func (sw *SWbot) cleanUpMap(links *map[string]time.Time) {
	for {
		for lichessId, timeAtStart := range *links {
			if time.Since(timeAtStart) > minLinkStayInMap {
				sw.mu.Lock()
				delete(*links, lichessId)
				sw.mu.Unlock()

			}
		}
	}
}
