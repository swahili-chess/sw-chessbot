package main

import (
	"time"

	"github.com/ChessSwahili/ChessSWBot/internal/data"
)

const (
	withGameIds = "&withGameIds=true"
	urlStatus   = "https://lichess.org/api/users/status?ids="
)

func (sw *SWbot) poller(listOfPlayerIdsChan <-chan []data.PlayerMinDt, listOfPlayerIds *[]data.PlayerMinDt) {

	ticker := time.NewTicker(time.Second * 6)

	defer ticker.Stop()

	go sw.cleanUpMap(sw.links)

	for range ticker.C {
		select {

		case playerIdsLists := <-listOfPlayerIdsChan:
			if len(playerIdsLists) != 0 {
				*listOfPlayerIds = playerIdsLists
			}

		default:
			url := prepareFetchInfoUrl(*listOfPlayerIds, urlStatus, withGameIds)

			if url != "" {
				sw.fetchPlayersInfo(url, sw.links)
			}
		}

	}
}
