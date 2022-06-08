package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
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
		MaxResults: 10,
	})
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
		id := t.Text[idIndex+len(idKeyword):idIndex+len(idKeyword)+8]
		// TODO: IDで既に送信済みか判定する
		fmt.Println(id)
		
		dateIndex := strings.Index(t.Text, dateKeyword)
		if dateIndex == -1 {
			fmt.Println("フォーマットが違う(期限)")
			continue
		}
		datetimeStr := t.Text[dateIndex+len(dateKeyword):dateIndex+len(dateKeyword)+11]
		datetime, _ := time.ParseInLocation(datetimeFormat, fmt.Sprintf("%d/%s", now.Year(), datetimeStr), jst)

		fmt.Println(now)
		fmt.Println(datetime)
		fmt.Println("---")
		if now.Before(datetime) {
			// TODO: メッセージを送信
			fmt.Println(t.Text)
		}
	}
}