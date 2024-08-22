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
	Username string `json:"name"`
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

		dt = append(dt, db.InsertLichessDataParams{
			LichessID: ctp.ID,
			Username:  ctp.Username,
		})

	}

	dt = append(dt, db.InsertLichessDataParams{
		LichessID: "herald18",
		Username:  "herald18",
	})

	return dt

}
