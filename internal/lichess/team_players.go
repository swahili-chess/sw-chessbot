package lichess

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	db "github.com/swahili-chess/sw-chessbot/internal/db/sqlc"
)

type TeamPlayer struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Perfs    struct {
		Chess960 struct {
			Games  int  `json:"games"`
			Rating int  `json:"rating"`
			Rd     int  `json:"rd"`
			Prog   int  `json:"prog"`
			Prov   bool `json:"prov"`
		} `json:"chess960"`
		Atomic struct {
			Games  int  `json:"games"`
			Rating int  `json:"rating"`
			Rd     int  `json:"rd"`
			Prog   int  `json:"prog"`
			Prov   bool `json:"prov"`
		} `json:"atomic"`
		RacingKings struct {
			Games  int  `json:"games"`
			Rating int  `json:"rating"`
			Rd     int  `json:"rd"`
			Prog   int  `json:"prog"`
			Prov   bool `json:"prov"`
		} `json:"racingKings"`
		UltraBullet struct {
			Games  int  `json:"games"`
			Rating int  `json:"rating"`
			Rd     int  `json:"rd"`
			Prog   int  `json:"prog"`
			Prov   bool `json:"prov"`
		} `json:"ultraBullet"`
		Blitz struct {
			Games  int  `json:"games"`
			Rating int  `json:"rating"`
			Rd     int  `json:"rd"`
			Prog   int  `json:"prog"`
			Prov   bool `json:"prov"`
		} `json:"blitz"`
		KingOfTheHill struct {
			Games  int  `json:"games"`
			Rating int  `json:"rating"`
			Rd     int  `json:"rd"`
			Prog   int  `json:"prog"`
			Prov   bool `json:"prov"`
		} `json:"kingOfTheHill"`
		Bullet struct {
			Games  int  `json:"games"`
			Rating int  `json:"rating"`
			Rd     int  `json:"rd"`
			Prog   int  `json:"prog"`
			Prov   bool `json:"prov"`
		} `json:"bullet"`
		Correspondence struct {
			Games  int  `json:"games"`
			Rating int  `json:"rating"`
			Rd     int  `json:"rd"`
			Prog   int  `json:"prog"`
			Prov   bool `json:"prov"`
		} `json:"correspondence"`
		Horde struct {
			Games  int  `json:"games"`
			Rating int  `json:"rating"`
			Rd     int  `json:"rd"`
			Prog   int  `json:"prog"`
			Prov   bool `json:"prov"`
		} `json:"horde"`
		Puzzle struct {
			Games  int  `json:"games"`
			Rating int  `json:"rating"`
			Rd     int  `json:"rd"`
			Prog   int  `json:"prog"`
			Prov   bool `json:"prov"`
		} `json:"puzzle"`
		Classical struct {
			Games  int  `json:"games"`
			Rating int  `json:"rating"`
			Rd     int  `json:"rd"`
			Prog   int  `json:"prog"`
			Prov   bool `json:"prov"`
		} `json:"classical"`
		Rapid struct {
			Games  int  `json:"games"`
			Rating int  `json:"rating"`
			Rd     int  `json:"rd"`
			Prog   int  `json:"prog"`
			Prov   bool `json:"prov"`
		} `json:"rapid"`
		Storm struct {
			Runs  int `json:"runs"`
			Score int `json:"score"`
		} `json:"storm"`
		Racer struct {
			Runs  int `json:"runs"`
			Score int `json:"score"`
		} `json:"racer"`
		Streak struct {
			Runs  int `json:"runs"`
			Score int `json:"score"`
		} `json:"streak"`
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
		Tv    int `json:"tv"`
	} `json:"playTime"`
	Title string `json:"title"`
}

func FetchTeamPlayers() []db.InsertLichessDataParams {

	var dt []db.InsertLichessDataParams
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://lichess.org/api/team/nyumbani-mates/users", nil)

	if err != nil {
		log.Error(err)

		return dt
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("LICHESS_TOKEN")))

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)

		return dt
	}

	defer resp.Body.Close()

	results := json.NewDecoder(resp.Body)

	for {

		var ctp TeamPlayer

		err := results.Decode(&ctp)

		if err != nil {
			if err != io.EOF {
				log.Error("we got an error while reading")
			}

			break
		}

		pd := db.InsertLichessDataParams{

			LichessID: ctp.ID,
			Username:  ctp.Username,
		}

		dt = append(dt, pd)

	}

	pd := db.InsertLichessDataParams{
		LichessID: "herald18",
		Username:  "herald18",
	}

	dt = append(dt, pd)

	return dt

}
