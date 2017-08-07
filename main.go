package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mmcdole/gofeed"
)

type feedConfig struct {
	Feeds []struct {
		Name string `json:"name"`
		Ivid string `json:"ivid"`
		URL  string `json:"url"`
	} `json:"feeds"`
	Tgtoken  string `json:"tgtoken"`
	Tgchatid int64  `json:"tgchatid"`
}

func loadConfig() feedConfig {
	buf := bytes.NewBuffer(nil)
	f, _ := os.Open("config.json")
	io.Copy(buf, f)
	f.Close()

	var jsonobject feedConfig

	err := json.Unmarshal(buf.Bytes(), &jsonobject)

	if err != nil {
		fmt.Println("error:", err)
	}

	return jsonobject
}

func loadLastTime() int64 {
	file, err := os.Open("lasttime")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	lastTime, _ := strconv.ParseInt(scanner.Text(), 10, 64)

	return lastTime
}

func saveLastTime() {
	file, err := os.Create("lasttime")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()
	fmt.Fprintf(file, "%d\n", time.Now().Unix())
}

func main() {
	for {
		config := loadConfig()

		lastTime := loadLastTime()

		bot, err := tgbotapi.NewBotAPI(config.Tgtoken)
		if err != nil {
			log.Panic(err)
		}

		for _, feedCfg := range config.Feeds {
			fp := gofeed.NewParser()
			feed, _ := fp.ParseURL(feedCfg.URL)
			for _, article := range feed.Items {

				articleTime := article.PublishedParsed.Unix()

				if articleTime > lastTime {
					text := fmt.Sprintf("https://t.me/iv?url=%s&rhash=%s", article.Link, feedCfg.Ivid)
					msg := tgbotapi.NewMessage(config.Tgchatid, text)
					bot.Send(msg)
				}
			}
		}
		saveLastTime()
		time.Sleep(time.Minute * 1)
	}
}
