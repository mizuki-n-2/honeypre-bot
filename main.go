package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/sivchari/gotwtr"
)

func Connection() redis.Conn {
	c, err := redis.DialURL(os.Getenv("REDIS_URL"), redis.DialTLSSkipVerify(true))
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func Exists(id string, c redis.Conn) (bool, error) {
	exists, err := redis.Bool(c.Do("EXISTS", id))
	return exists, err
}

func Set(key string, value string, c redis.Conn) error {
	if _, err := c.Do("SET", key, value); err != nil {
		return err
	}

	// 有効期限は1時間
	if _, err := c.Do("EXPIRE", key, 60*60); err != nil {
		return err
	}

	return nil
}

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

	bot, err := linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	c := Connection()
	defer c.Close()

	idKeyword := "ゲリラ招待ID:"
	dateKeyword := "期限:"

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatal(err)
	}
	now := time.Now().In(jst)
	datetimeFormat := "2006/01/02 15:04"

	for _, t := range tsr.Tweets {
		// Textから日時を抽出してフォーマット
		dateIndex := strings.Index(t.Text, dateKeyword)
		if dateIndex == -1 {
			fmt.Println("フォーマットが違う(期限)")
			continue
		}
		datetimeStr := t.Text[dateIndex+len(dateKeyword) : dateIndex+len(dateKeyword)+11]
		// yearは取得できないので、実行時と同じ年を設定
		datetime, _ := time.ParseInLocation(datetimeFormat, fmt.Sprintf("%d/%s", now.Year(), datetimeStr), jst)

		if now.Before(datetime) {
			// TextからIDを抽出
			idIndex := strings.Index(t.Text, idKeyword)
			if idIndex == -1 {
				fmt.Println("フォーマットが違う(ゲリラ招待ID)")
				continue
			}
			id := t.Text[idIndex+len(idKeyword) : idIndex+len(idKeyword)+8]

			// IDで既に送信済みか判定する
			isExists, err := Exists(id, c)
			if err != nil {
				log.Fatal(err)
			}
			if isExists {
				fmt.Println("既に送信済み")
				continue
			}

			fmt.Println(t.Text) // ログ用

			// メッセージを送信する
			if _, err := bot.PushMessage(os.Getenv("MY_USER_ID"), linebot.NewTextMessage(t.Text)).Do(); err != nil {
				log.Fatal(err)
			}

			// IDをRedisに保存する
			if err := Set(id, "", c); err != nil {
				log.Fatal(err)
			}
		}
	}
}
