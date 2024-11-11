package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
	"github.com/swahili-chess/sw-chessbot/internal/lichess"
	"github.com/swahili-chess/sw-chessbot/internal/poll"
	"github.com/swahili-chess/sw-chessbot/internal/req"
)

const (
	start_txt       = "Use this bot to get link of games of Chesswahili team members that are actively playing on Lichess. Type /stop to stop receiving notifications`"
	stop_txt        = "Sorry to see you leave You wont be receiving notifications. Type /start to receive"
	unknown_cmd     = "I don't know that command"
	maintenance_txt = "We are having Bot maintenance. Service will resume shortly"
)

func init() {

	var programLevel = new(slog.LevelVar)
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))

}

type UpdateTgBotUsersParams struct {
	Isactive bool  `json:"isactive"`
	ID       int64 `json:"id"`
}

type InsertTgBotUsersParams struct {
	ID       int64 `json:"id"`
	Isactive bool  `json:"isactive"`
}

func main() {

	var is_maintenance_txt = false
	var API_URL string
	var botToken string

	flag.StringVar(&API_URL, "db-dsn", os.Getenv("API_URL"), "API URL")
	flag.StringVar(&botToken, "bot-token", os.Getenv("TG_BOT_TOKEN"), "Bot Token")

	flag.Parse()

	if botToken == "" || API_URL == "" {
		slog.Error("Bot token or DSN not provided")
		return
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		slog.Error("failed to create bot api instance", "err", err)
		return
	}

	links := make(map[string]time.Time)

	swbot := &poll.SWbot{
		Bot:   bot,
		Links: &links,
	}

	u := tgbotapi.NewUpdate(0)

	u.Timeout = 60

	membersIdsChan := make(chan []lichess.InsertMemberParams)

	updates := bot.GetUpdatesChan(u)

	//Fetch from the team for the first time
	memberIds := lichess.FetchTeamMembers()
	if len(memberIds) == 0 {
		slog.Error("length of player ids shouldn't be 0")
	}
	swbot.InsertNewMembers(memberIds)

	go swbot.PollTeam(membersIdsChan)

	go swbot.PollMember(membersIdsChan, &memberIds)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !update.Message.IsCommand() {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		switch update.Message.Command() {
		case "start":
			msg.Text = start_txt
			botUser := InsertTgBotUsersParams{
				ID:       update.Message.From.ID,
				Isactive: true,
			}
			var errResponse req.ErrorResponse

			statusCode, err := req.PostOrPutRequest(http.MethodPost, "https://api.swahilichess.com/telegram/bot/users", botUser, &errResponse)
			if statusCode == http.StatusInternalServerError {
				switch {
				case errResponse.Error == `pq: duplicate key value violates unique constraint "tgbot_users_pkey"`:
					args := UpdateTgBotUsersParams{
						ID:       botUser.ID,
						Isactive: botUser.Isactive,
					}

					statusCode, err := req.PostOrPutRequest(http.MethodPut, "https://api.swahilichess.com/telegram/bot/users", args, &errResponse)
					if statusCode == http.StatusInternalServerError {
						slog.Error("failed to update bot user", "error", errResponse.Error)
					} else if err != nil {
						slog.Error("failed to update bot user", "error", err)
					}

				default:
					slog.Error("failed to insert bot user", "err", err)
				}
			} else if err != nil {
				slog.Error("failed to insert bot user", "error", err, "statuscode", statusCode)
			}

		case "stop":
			botUser := UpdateTgBotUsersParams{
				ID:       update.Message.From.ID,
				Isactive: false,
			}
			var errResponse req.ErrorResponse

			statusCode, err := req.PostOrPutRequest(http.MethodPut, "https://api.swahilichess.com/telegram/bot/users", botUser, &errResponse)
			if statusCode == http.StatusInternalServerError {
				slog.Error("failed to update bot user", "error", errResponse.Error)
			} else if err != nil {
				slog.Error("failed to update bot user", "error", err)
			}
			msg.Text = stop_txt

		case "subs":
			var res []int64
			var errResponse req.ErrorResponse
			statusCode, err := req.GetRequest("https://api.swahilichess.com/telegram/bot/users/active", &res, &errResponse)
			if statusCode != http.StatusInternalServerError {
				slog.Error("failed to get telegram bot users", "err", errResponse.Error)

			} else if statusCode != http.StatusOK || err != nil {
				slog.Error("failed to get telegram bot users", "err", err, "statusCode", statusCode)
			}
			msg.Text = fmt.Sprintf("There are %d subscribers in chesswahiliBot", len(res))

		case "ml":
			msg.Text = fmt.Sprintf("There are %d in a map so far.", len(*swbot.Links))

		case "sm":
			if poll.Master_ID == update.Message.From.ID {
				is_maintenance_txt = true
			}

		case "help":
			msg.Text = `
			Commands for this @chesswahiliBot bot are:
			
			/start  start the bot (i.e., enable receiving of the game links)
			/stop   stop the bot (i.e., disable receiving of the game links)
			/subs   subscribers for the bot
			/ml     current map length
			/help   this help text
			/sm     send maintenace message for @Hopertz only.
			`

		default:
			msg.Text = unknown_cmd
		}

		if is_maintenance_txt {
			swbot.SendMaintananceMsg(maintenance_txt)
			is_maintenance_txt = false

		} else {
			if _, err := swbot.Bot.Send(msg); err != nil {
				slog.Error("failed to send msg", "err", err, "msg", msg)
			}
		}

	}
}
