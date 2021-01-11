package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/wbergg/bordershop-bot/tele"
)

type Request struct {
	ProductID string   `json:"productId"`
	SiteIds   []string `json:"siteIds"`
}

type Response []struct {
	SiteID            string `json:"SiteId"`
	StockTextShort    string `json:"StockTextShort"`
	StockTextLong     string `json:"StockTextLong"`
	ShowStock         bool   `json:"ShowStock"`
	SectionLabel      string `json:"SectionLabel"`
	ShelfLabel        string `json:"ShelfLabel"`
	Shelf             string `json:"Shelf"`
	Section           string `json:"Section"`
	NotYetSaleStarted string `json:"NotYetSaleStarted"`
	IsAgent           bool   `json:"IsAgent"`
	TranslateService  bool   `json:"TranslateService"`
}

func main() {
	// Enable bool debug flag
	debug := flag.Bool("debug", false, "Turns on debug mode and prints to stdout")

	//Telegram API key
	tgAPIKey := os.Getenv("AB_APIKEY")
	if tgAPIKey == "" {
		panic("No valid Telegram API Key specified")
	}
	//Telegram channel number
	tgChannel, _ := strconv.ParseInt(os.Getenv("AB_CHANNEL"), 10, 64)
	if tgChannel == 0 {
		panic("No valid Telegram channel specified")
	}

	tg := tele.New(tgAPIKey, tgChannel, false, *debug)
	tg.Init(false)

	// Run before starting the timer
	requestData(tg)

	// Set poll interval to 5 minutes
	pollInterval := 5

	tmr := time.Tick(time.Duration(pollInterval) * time.Minute)
	// Loop forever
	for range tmr {
		requestData(tg)
	}
}

func requestData(t *tele.Tele) {
	var ctx = context.Background()

	// Redis initialization
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	s := Response{}
	jsonValue, err := json.Marshal(Request{
		// Product ID from Systembolaget
		ProductID: "508393",
		// Site (store) ID(s) from Systembolaget
		SiteIds: []string{"0611"},
	})
	if err != nil {
		panic(err)
	}

	// Get stock balance
	res, err := http.Post("https://www.systembolaget.se/api/product/getstockbalance",
		"application/json",
		bytes.NewBuffer(jsonValue))
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	json.Unmarshal(body, &s)

	//Prepare Telegram message
	message := ""
	message = message + "*DEKADENS UPDATE 2000!*\n"
	message = message + "https://static.systembolaget.se/imagelibrary/publishedmedia/45ve1hzsi2adzw1b7m4v/508393.jpg" + "\n\n"
	message = message + "Someone just bought Arboga 10.2 @ Systembolaget in Gislaved!\n\n"

	dataUpdate := false

	for _, site := range s {

		val, err := rdb.Get(ctx, site.SiteID).Result()
		if err != nil {
			panic(err)
		}

		if val == site.StockTextShort {
			fmt.Println("No stock update, currently at " + site.StockTextShort)
		} else {
			err := rdb.Set(ctx, site.SiteID, site.StockTextShort, 0).Err()
			if err != nil {
				panic(err)
			}
			message = message + "Currently left in stock: " + site.StockTextShort + "\n"
			dataUpdate = true
		}

	}
	if dataUpdate == true {
		// Send message to Telegram
		t.SendM(message)
		fmt.Println(message)
	}
}
