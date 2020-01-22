package main

import (
	"bytes"
	"encoding/json"
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
	//Telegram API key
	api_key := os.Getenv("AB_APIKEY")
	if api_key == "" {
		panic("No valid Telegram API Key")
	}
	//Telegram channel number
	channel, _ := strconv.ParseInt(os.Getenv("AB_CHANNEL"), 10, 64)

	tg := tele.New(api_key, channel, false)
	tg.Init()

	// Run before starting the timer
	requestData(tg)

	tmr := time.NewTimer(5 * time.Minute)

	for {
		select {
		case <-tmr.C:
			requestData(tg)
		}
	}
}

func requestData(t *tele.Tele) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}

	s := Response{}
	jsonValue, err := json.Marshal(Request{
		ProductID: "508393",
		SiteIds:   []string{"0611"},
	})
	if err != nil {
		panic(err)
	}

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

		val, err := client.Get(site.SiteID).Result()
		if err != nil {
			panic(err)
		}

		if val == site.StockTextShort {
			fmt.Println("No stock update, currently at " + site.StockTextShort)
		} else {
			err = client.Set(site.SiteID, site.StockTextShort, 0).Err()
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
