package main

import (
	"time"
)

func (sw *SWbot) poller(usersId []string) {

	ticker := time.NewTicker(time.Second * 6)

	defer ticker.Stop()

	url := prepareUrl(usersId)

	go sw.cleanUpMap(sw.links)

	for range ticker.C {

		sw.fetchStatus(url, sw.links)

	}
}
