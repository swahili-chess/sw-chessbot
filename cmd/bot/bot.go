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

func main() {

	var is_maintenance_txt = false
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
		return
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

	membersIdsChan := make(chan []db.InsertMemberParams)

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
			msg.Text = stop_txt
		case "subs":
			res, err := swbot.Store.GetActiveTgBotUsers(context.Background())
			if err != nil {
				slog.Error("failed to get bot active members", "err", err)
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

func openDB(dsn string) (*sql.DB, error) {

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return db, db.PingContext(ctx)
}
