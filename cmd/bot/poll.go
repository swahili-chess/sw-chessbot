package main

import (
	"time"
)

type poller struct {
	ticker *time.Ticker // periodic ticker
	url    string
}

func newPoller() *poller {
	rv := &poller{
		ticker: time.NewTicker(time.Second * 6),
		url:    "https://lichess.org/api/users/status?ids=",
	}

	go rv.run()
	return rv
}

func (p *poller) run() {

	for range p.ticker.C {

		fetch(p.url)

	}
}

func fetch(url string) {
	// do some fetch
}
