package main

import (
	"time"
)

func (sw *SWbot) poller(usersId []string) {

	ticker := time.NewTicker(time.Second * 6)

	defer ticker.Stop()

	url := prepareUrl(usersId)

	links := make(map[string]bool)

	for range ticker.C {

		sw.fetchStatus(url, links)

	}
}
