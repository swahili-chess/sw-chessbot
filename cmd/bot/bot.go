package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"

	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
	db "github.com/swahili-chess/sw-chessbot/internal/db/sqlc"
	"github.com/swahili-chess/sw-chessbot/internal/lichess"
	"github.com/swahili-chess/sw-chessbot/internal/poll"
)

const startTxt = "Use this bot to get link of games of Chesswahili team members that are actively playing on Lichess. Type /stop to stop receiving notifications`"

const stopTxt = "Sorry to see you leave You wont be receiving notifications. Type /start to receive"

const dontTxt = "I don't know that command"

const masterID = 731217828

var maintanenanceTxT = "We are having Bot maintenance. Service will resume shortly"

var IsMaintananceCost = false

func init() {

	var programLevel = new(slog.LevelVar)

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))

}

func main() {
	var dsn string
	var botToken string

	flag.StringVar(&dsn, "db-dsn", os.Getenv("DSN_BOT"), "Postgres DSN")
	flag.StringVar(&botToken, "bot-token", os.Getenv("TG_BOT_TOKEN"), "Bot Token")

	flag.Parse()

	if botToken == "" || dsn == "" {
		slog.Error("Bot token or DSN not provided")
		return
	}

	con, err := openDB(dsn)

	if err != nil {
		slog.Error("failed connect to db", "err", err)
	}

	defer con.Close()

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		slog.Error("failed to create bot api instance", "err", err)
		return
	}

	links := make(map[string]time.Time)

	swbot := &poll.SWbot{
		Bot:   bot,
		Links: &links,
		Store: db.NewStore(con),
	}

	u := tgbotapi.NewUpdate(0)

	u.Timeout = 60

	listOfPlayerIdsChan := make(chan []db.InsertLichessDataParams)

	updates := bot.GetUpdatesChan(u)

	//Fetch player ids from the team for the first time
	listOfPlayerIds := lichess.FetchTeamPlayers()

	if len(listOfPlayerIds) == 0 {
		slog.Error("length of player ids shouldn't be 0")
	}

	swbot.InsertUsernames(listOfPlayerIds)

	// Fetch player  ids after in the team after every 5 minutes
	go swbot.PollTeam(listOfPlayerIdsChan)

	go swbot.Poller(listOfPlayerIdsChan, &listOfPlayerIds)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		// Create a new MessageConfig. We don't have text yet,
		// so we leave it empty.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// Extract the command from the Message
		switch update.Message.Command() {
		case "start":
			msg.Text = startTxt
			botUser := db.InsertTgBotUsersParams{
				ID:       update.Message.From.ID,
				Isactive: true,
			}
			err := swbot.Store.InsertTgBotUsers(context.Background(), botUser)

			if err != nil {
				switch {
				case err.Error() == `pq: duplicate key value violates unique constraint "tgbot_users_pkey"`:
					args := db.UpdateTgBotUsersParams{
						ID:       botUser.ID,
						Isactive: botUser.Isactive,
					}
					err := swbot.Store.UpdateTgBotUsers(context.Background(), args)
					if err != nil {
						slog.Error("failed to update bot user", "err", err, "args", args)
					}

				default:
					slog.Error("failed to insert bot user", "err", err, "args", botUser)
				}
			}

		case "stop":
			botUser := db.UpdateTgBotUsersParams{
				ID:       update.Message.From.ID,
				Isactive: false,
			}
			err := swbot.Store.UpdateTgBotUsers(context.Background(), botUser)
			if err != nil {
				slog.Error("failed to update bot user", "err", err, "args", botUser)
			}
			msg.Text = stopTxt
		case "subs":
			res, err := swbot.Store.GetActiveTgBotUsers(context.Background())
			if err != nil {
				slog.Error("failed to get bot active members", "err", err)
			}
			msg.Text = fmt.Sprintf("There are %d subscribers in chesswahiliBot", len(res))

		case "ml":
			msg.Text = fmt.Sprintf("There are %d in a map so far.", len(*swbot.Links))

		case "sm":
			if masterID == update.Message.From.ID {
				IsMaintananceCost = true
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
			msg.Text = dontTxt
		}

		if IsMaintananceCost {
			swbot.SendMaintananceMsg(maintanenanceTxT)
			IsMaintananceCost = false

		} else {
			if _, err := swbot.Bot.Send(msg); err != nil {
				slog.Error("failed to send msg", "err", err, "msg", msg)
			}
		}

	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	db.PingContext(ctx)

	return db, err
}
