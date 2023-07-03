package main

import (
	"time"
)

func (sw *SWbot) poller(listOfPlayerIdsChan <-chan []string, listOfPlayerIds *[]string) {

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
			url := prepareFetchInfoUrl(*listOfPlayerIds)
			sw.fetchPlayersInfo(url, sw.links)
		}

	}
}
