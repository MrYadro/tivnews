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

func loadLastTime(hash string) int64 {
	file, err := os.Open(hash)
	if err != nil {
		saveLastTime(hash, time.Now())
		log.Print(err)
		return time.Now().Unix()
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	lastTime, _ := strconv.ParseInt(scanner.Text(), 10, 64)

	return lastTime
}

func saveLastTime(hash string, timer time.Time) {
	file, err := os.Create(hash)
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()
	fmt.Fprintf(file, "%d\n", timer.Unix())
}

func main() {
	for {
		config := loadConfig()

		bot, err := tgbotapi.NewBotAPI(config.Tgtoken)
		if err != nil {
			continue
		}

		for _, feedCfg := range config.Feeds {
			fp := gofeed.NewParser()
			feed, err := fp.ParseURL(feedCfg.URL)
			if err != nil {
				continue
			}
			ivid := feedCfg.Ivid
			lastTime := loadLastTime(ivid)
			for i, article := range feed.Items {
				if i == 0 {
					saveLastTime(ivid, *article.PublishedParsed)
				}
				articleTime := article.PublishedParsed.Unix()

				if articleTime > lastTime {
					text := fmt.Sprintf("Visit article page at [%s](https://t.me/iv?url=%s&rhash=%s)", feedCfg.Name, article.Link, ivid)
					msg := tgbotapi.NewMessage(config.Tgchatid, text)
					msg.ParseMode = tgbotapi.ModeMarkdown
					bot.Send(msg)
				}
			}
		}
		time.Sleep(time.Minute * 5)
	}
}
