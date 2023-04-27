package main

import (
	"time"
)

func (sw *SWbot) poller(playersIdChan <-chan []string, playersId *[]string) {

	ticker := time.NewTicker(time.Second * 6)

	defer ticker.Stop()

	go sw.cleanUpMap(sw.links)

	for range ticker.C {
		select {
		case playersChanValue := <-playersIdChan:
			*playersId = playersChanValue

		default:
			url := prepareFetchStatusUrl(*playersId)
			sw.fetchPlayersStatus(url, sw.links)
		}

	}
}
