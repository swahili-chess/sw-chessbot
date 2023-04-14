package main

import (
	"time"
)

func (sw *SWbot) poller(playersId []string) {

	ticker := time.NewTicker(time.Second * 6)

	defer ticker.Stop()

	url := prepareFetchStatusUrl(playersId)

	go sw.cleanUpMap(sw.links)

	for range ticker.C {

		sw.fetchPlayersStatus(url, sw.links)

	}
}
