package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	searchURL = "http://ano.lolcathost.org/json/tag.json"
	randomURL = "http://ano.lolcathost.org/json/pic.json"
	picsURL   = "http://ano.lolcathost.org/pics"
	thumbsURL = "http://ano.lolcathost.org/thumbs/s"
	picsLimit = 25
)

func main() {
	parallel := flag.Int("parallel", 10, "maximum number of parallel goroutines")
	debug := flag.Bool("debug", false, "enable debug output")
	flag.Usage = usage
	flag.Parse()

	token := os.Getenv("ANOLOLCATBOT_TOKEN")
	if token == "" {
		log.Fatal("missing env var ANOLOLCATBOT_TOKEN")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("could not create bot API: %v", err)
	}

	bot.Debug = *debug

	log.Printf("authorized on account %s", bot.Self.UserName)

	// Get only the last remaining update.
	// Reference: https://core.telegram.org/bots/api#getupdates
	u := tgbotapi.NewUpdate(-1)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("could not get updates channel: %v", err)
	}

	parchan := make(chan struct{}, *parallel)

	for update := range updates {
		if update.InlineQuery == nil {
			continue
		}

		log.Printf("new update: %v - %q", update.InlineQuery.From, update.InlineQuery.Query)

		parchan <- struct{}{}
		go handleUpdate(parchan, bot, update)
	}
}

func handleUpdate(parchan <-chan struct{}, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	defer func() {
		<-parchan
	}()

	q := strings.TrimSpace(update.InlineQuery.Query)
	id := update.InlineQuery.ID
	offset := update.InlineQuery.Offset
	nextOffset := ""

	var (
		err     error
		results []interface{}
	)

	if q != "" {
		offsetInt := 0

		if offset != "" {
			offsetInt, err = strconv.Atoi(offset)
			if err != nil {
				log.Printf("cannot parse offset: %q", offset)
				return
			}
		}

		results, err = searchRelated(q, offsetInt)
		if err != nil {
			log.Printf("cannot get results: %v", err)
			return
		}

		if len(results) == picsLimit {
			nextOffset = strconv.Itoa(offsetInt + picsLimit)
		}
	} else {
		results, err = randomPics()
		if err != nil {
			log.Printf("cannot get results: %v", err)
			return
		}
		nextOffset = "random"
	}

	cfg := tgbotapi.InlineConfig{
		InlineQueryID: id,
		Results:       results,
		NextOffset:    nextOffset,
	}

	if _, err := bot.AnswerInlineQuery(cfg); err != nil {
		log.Printf("cannot answer inline query: %v", err)
	}
}

func searchRelated(query string, offset int) (results []interface{}, err error) {
	reqData := struct {
		Method string   `json:"method"`
		Tags   []string `json:"tags"`
		Offset int      `json:"offset"`
		Limit  int      `json:"limit"`
	}{
		"searchRelated",
		strings.Split(query, ","),
		offset,
		picsLimit,
	}

	return getResults(searchURL, reqData)
}

func randomPics() (results []interface{}, err error) {
	reqData := struct {
		Method string `json:"method"`
		Num    int    `json:"num"`
	}{
		"random",
		picsLimit,
	}

	return getResults(randomURL, reqData)
}

func getResults(url string, reqData interface{}) (results []interface{}, err error) {
	respData := struct {
		Pics []struct {
			ID     string `json:"id"`
			UID    string `json:"uid"`
			Width  int    `json:"sizew"`
			Height int    `json:"sizeh"`
		}
	}{}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("cannot encode request: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("cannot send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code is not 200: %v", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return nil, fmt.Errorf("cannot decode response: %v", err)
	}

	if len(respData.Pics) == 0 {
		return nil, nil
	}

	for _, p := range respData.Pics {
		r := newInlineQueryResult(p.UID, p.ID, p.Width, p.Height)
		results = append(results, r)
	}

	return results, nil
}

func newInlineQueryResult(uid, id string, width, height int) interface{} {
	var ret interface{}

	if len(uid) > 64 {
		uid = uid[:64]
	}

	ext := strings.ToLower(filepath.Ext(id))
	if ext == ".gif" {
		gif := tgbotapi.NewInlineQueryResultGIF(uid, fmt.Sprintf("%v/%v", picsURL, id))
		gif.ThumbURL = fmt.Sprintf("%v/%v", thumbsURL, id)
		gif.Width = width
		gif.Height = height
		ret = gif
	} else {
		photo := tgbotapi.NewInlineQueryResultPhoto(uid, fmt.Sprintf("%v/%v", picsURL, id))
		photo.ThumbURL = fmt.Sprintf("%v/%v", thumbsURL, id)
		photo.Width = width
		photo.Height = height
		ret = photo
	}

	return ret
}

func usage() {
	fmt.Println("usage: anololcatbot [opts]")
	flag.PrintDefaults()
}
