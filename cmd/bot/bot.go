package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"time"

	"github.com/ChessSwahili/ChessSWBot/internal/data"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

func main() {
	var dsn string

	flag.StringVar(&dsn, "db-dsn", os.Getenv("DSN_BOT"), "Postgres DSN")

	flag.Parse()

	db, err := openDB(dsn)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	models := data.NewModels(db)

	bot, err := tgbotapi.NewBotAPI(os.Getenv("Token"))
	if err != nil {
		log.Panic(err)
	}


	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

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
			msg.Text = "Please use this bot to get link of games of Chesswahili team members that are actively playing on Lichess. Type /stop to stop receiving notifications`;"
			botUser := &data.User{
				ID:       update.Message.From.ID,
				Isactive: true,
			}
			err := models.Users.Insert(botUser)

			if err != nil {
				log.Println(err)
			}

		case "stop":
			msg.Text = "Sorry to see you leave You wont be receiving notifications. Type /start to receive "
		case "team":
			msg.Text = "There are 10 members in chesswahili"
		default:
			msg.Text = "I don't know that command"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
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
