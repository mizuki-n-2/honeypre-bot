package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/sivchari/gotwtr"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("読み込み出来ませんでした: %v", err)
	}

	token := os.Getenv("BEARER_TOKEN")
	client := gotwtr.New(token)
	tsr, err := client.SearchRecentTweets(context.Background(), "ハニプレ ゲリラライブ Lv3", &gotwtr.SearchTweetsOption{
		MediaFields: []gotwtr.MediaField{gotwtr.MediaFieldURL},
		MaxResults:  10,
	})
	if err != nil {
		log.Fatal(err)
	}

	channelSecret := os.Getenv("CHANNEL_SECRET")
	channelToken := os.Getenv("CHANNEL_TOKEN")
	bot, err := linebot.New(channelSecret, channelToken)
	if err != nil {
		log.Fatal(err)
	}

	idKeyword := "ゲリラ招待ID:"
	dateKeyword := "期限:"
	now := time.Now()
	datetimeFormat := "2006/01/02 15:04"
	jst, _ := time.LoadLocation("Asia/Tokyo")

	for _, t := range tsr.Tweets {
		idIndex := strings.Index(t.Text, idKeyword)
		if idIndex == -1 {
			fmt.Println("フォーマットが違う(ゲリラ招待ID)")
			continue
		}
		id := t.Text[idIndex+len(idKeyword) : idIndex+len(idKeyword)+8]
		// TODO: IDで既に送信済みか判定する
		fmt.Println(id)

		dateIndex := strings.Index(t.Text, dateKeyword)
		if dateIndex == -1 {
			fmt.Println("フォーマットが違う(期限)")
			continue
		}
		datetimeStr := t.Text[dateIndex+len(dateKeyword) : dateIndex+len(dateKeyword)+11]
		currentYear := now.Year()
		if t.CreatedAt != "" {
			createdAt, _ := time.ParseInLocation(datetimeFormat, t.CreatedAt, jst)
			currentYear = createdAt.Year()
		}
		datetime, _ := time.ParseInLocation(datetimeFormat, fmt.Sprintf("%d/%s", currentYear, datetimeStr), jst)

		if now.Before(datetime) {
			fmt.Println(t.Text)
			if _, err := bot.PushMessage(os.Getenv("MY_USER_ID"), linebot.NewTextMessage(t.Text)).Do(); err != nil {
				log.Fatal(err)
			}
		}
	}
}
