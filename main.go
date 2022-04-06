package main

import (
	"context"
	"fmt"
	"log"
	"os"

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
	
	for _, t := range tsr.Tweets {
		fmt.Println("---")
		fmt.Println(t.Text)
	}
}