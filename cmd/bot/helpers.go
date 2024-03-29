package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ChessSwahili/ChessSWBot/internal/data"
	log "github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
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

// Send the link to the telegram ids
func (sw *SWbot) sendMsgToTelegramIds(linkId string) {
	gameLink := base_url + linkId

	ids, _ := sw.models.Users.GetActiveUsers()

	for _, id := range ids {
		msg := tgbotapi.NewMessage(id.Id, gameLink)

		sw.bot.Send(msg)
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
		log.Error("Error while fetching status")
		return
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&listOfPlayerInfos)

	if err != nil {
		log.Error("Error decoding the json body", err)
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

				sw.sendMsgToTelegramIds(playerInfo.PlayingId)
			}

		}
	}

}

// Prepare the url to fetch the status of the players
func prepareFetchInfoUrl(players []data.PlayerMinDt, urlStatus, withGameIds string) string {

	if len(players) == 0 {
		return ""
	}
	playersIds := []string{}

	for _, player := range players {
		playersIds = append(playersIds, player.ID)
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

// Send a message to all active users when the bot is down for maintanance
func (sw *SWbot) sendMaintananceMsg(msg string) {

	ids, _ := sw.models.Users.GetActiveUsers()

	for _, id := range ids {
		msg := tgbotapi.NewMessage(id.Id, msg)

		sw.bot.Send(msg)
	}
}

func (sw *SWbot) InsertUsernames(list []data.PlayerMinDt) {
	// get current usernames in db
	lichess_ids, err := sw.models.Lichess.GetLichessUsernames()

	if err != nil {
		log.Error("Failed to get usernames in DB")
		return
	}

	newPlayers := findNewPlayers(lichess_ids, list)

	for _, player := range newPlayers {

		err := sw.models.Lichess.Insert(player)

		if err != nil {
			log.Error("Failed to insert user", player)
		}

	}
}

func findNewPlayers(lichess_ids []string, players []data.PlayerMinDt) []data.PlayerMinDt {
	newPlayers := []data.PlayerMinDt{}
	elementSet := make(map[string]bool)

	for _, lichess_id := range lichess_ids {
		elementSet[lichess_id] = true
	}

	for _, dt := range players {
		if _, found := elementSet[dt.ID]; !found {
			newPlayers = append(newPlayers, dt)
		} else {
			delete(elementSet, dt.ID) // Remove common elements
		}
	}

	return newPlayers
}
