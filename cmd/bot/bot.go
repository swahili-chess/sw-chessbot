package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ChessSwahili/ChessSWBot/internal/data"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

const startTxt = "Use this bot to get link of games of Chesswahili team members that are actively playing on Lichess. Type /stop to stop receiving notifications`"

const stopTxt = "Sorry to see you leave You wont be receiving notifications. Type /start to receive"

const dontTxt = "I don't know that command"

const masterID = 731217828

var maintanenanceTxT = "We are having Bot maintenance. Service will resume shortly"

var IsMaintananceCost = false

type SWbot struct {
	bot    *tgbotapi.BotAPI
	models data.Models
	links  *map[string]time.Time
	mu     sync.RWMutex
}

func init() {

	log.SetFormatter(&log.JSONFormatter{})

	log.SetOutput(os.Stdout)

	log.SetLevel(log.InfoLevel)
}

func main() {
	var dsn string
	var botToken string

	flag.StringVar(&dsn, "db-dsn", os.Getenv("DSN_BOT"), "Postgres DSN")
	flag.StringVar(&botToken, "bot-token", os.Getenv("TG_BOT_TOKEN"), "Bot Token")

	flag.Parse()

	if botToken == "" || dsn == "" {
		log.Fatal("Bot token or DSN not provided")
	}

	db, err := openDB(dsn)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	models := data.NewModels(db)

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	links := make(map[string]time.Time)
	swbot := SWbot{
		bot:    bot,
		models: models,
		links:  &links,
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	listOfPlayerIdsChan := make(chan []string)

	updates := bot.GetUpdatesChan(u)

	//Fetch player ids from the team for the first time
	listOfPlayerIds := data.FetchTeamPlayers()

	if len(listOfPlayerIds) == 0 {
		log.Fatal("No player ids found")
	}

	go swbot.InsertUsernames(listOfPlayerIds)

	// Fetch player  ids after in the team after every 5 minutes
	go swbot.pollTeam(listOfPlayerIdsChan)

	go swbot.poller(listOfPlayerIdsChan, &listOfPlayerIds)

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
			botUser := &data.User{
				ID:       update.Message.From.ID,
				Isactive: true,
			}
			err := swbot.models.Users.Insert(botUser)

			if err != nil {
				switch {
				case err.Error() == `pq: duplicate key value violates unique constraint "users_pkey"`:

					err := swbot.models.Users.Update(botUser)
					if err != nil {
						log.Error(err)
					}

				default:
					log.Error(err)
				}
			}

		case "stop":
			botUser := &data.User{
				ID:       update.Message.From.ID,
				Isactive: false,
			}
			err := swbot.models.Users.Update(botUser)
			if err != nil {
				log.Error(err)
			}
			msg.Text = stopTxt
		case "subs":
			res, err := models.Users.GetActiveUsers()
			if err != nil {
				log.Error(err)
			}
			msg.Text = fmt.Sprintf("There are %d subscribers in chesswahiliBot", len(res))

		case "ml":
			msg.Text = fmt.Sprintf("There are %d in a map so far.", len(*swbot.links))

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
			swbot.sendMaintananceMsg(maintanenanceTxT)
			IsMaintananceCost = false

		} else {
			if _, err := swbot.bot.Send(msg); err != nil {
				log.Error(err)
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

	if err != nil {
		return nil, err
	}
	return db, err
}
