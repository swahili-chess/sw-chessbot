package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type ChessTeamPlayer struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Perfs    map[string]struct {
		Games  int  `json:"games"`
		Rating int  `json:"rating"`
		Rd     int  `json:"rd"`
		Prog   int  `json:"prog"`
		Prov   bool `json:"prov"`
		Runs   int  `json:"runs,omitempty"`
		Score  int  `json:"score,omitempty"`
	} `json:"perfs"`

	CreatedAt    int64 `json:"createdAt"`
	Disabled     bool  `json:"disabled"`
	TosViolation bool  `json:"tosViolation"`
	Profile      struct {
		Country    string `json:"country"`
		Location   string `json:"location"`
		Bio        string `json:"bio"`
		FirstName  string `json:"firstName"`
		LastName   string `json:"lastName"`
		FideRating int    `json:"fideRating"`
		UscfRating int    `json:"uscfRating"`
		EcfRating  int    `json:"ecfRating"`
		Links      string `json:"links"`
	} `json:"profile"`
	SeenAt   int64 `json:"seenAt"`
	Patron   bool  `json:"patron"`
	Verified bool  `json:"verified"`
	PlayTime struct {
		Total int `json:"total"`
		TV    int `json:"tv"`
	} `json:"playTime"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Playing string `json:"playing"`
	Count   struct {
		All      int `json:"all"`
		Rated    int `json:"rated"`
		AI       int `json:"ai"`
		Draw     int `json:"draw"`
		DrawH    int `json:"drawH"`
		Loss     int `json:"loss"`
		LossH    int `json:"lossH"`
		Win      int `json:"win"`
		WinH     int `json:"winH"`
		Bookmark int `json:"bookmark"`
		Playing  int `json:"playing"`
		Import   int `json:"import"`
		Me       int `json:"me"`
	} `json:"count"`
	Streaming  bool `json:"streaming"`
	Followable bool `json:"followable"`
	Following  bool `json:"following"`
	Blocking   bool `json:"blocking"`
	FollowsYou bool `json:"followsYou"`
}

func FetchTeamPlayers() []string {

	var ids []string
	client := http.DefaultClient

	req, err := http.NewRequest("GET", "https://lichess.org/api/team/nyumbani-mates/users", nil)

	if err != nil {
		fmt.Println(err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("LICHESS_TOKEN")))

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	results := json.NewDecoder(resp.Body)

	for {

		var ctp ChessTeamPlayer

		err := results.Decode(&ctp)

		if err != nil {
			if err != io.EOF {
				fmt.Println("fuck we got an error while reading")
			}

			break
		}

		ids = append(ids, ctp.ID)

	}
	ids = append(ids, "herald18")
	return ids

}
