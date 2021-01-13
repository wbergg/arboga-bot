package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/wbergg/bordershop-bot/tele"
)

type Request struct {
	ProductID string   `json:"productId"`
	StoreID   []string `json:"storeId"`
}

type Response []struct {
	ProductID string `json:"productId"`
	StoreID   string `json:"storeId"`
	Shelf     string `json:"shelf"`
	Stock     int    `json:"stock"`
}

func main() {
	// Enable bool debug flag
	debug := flag.Bool("debug", false, "Turns on debug mode and prints to stdout")

	//Telegram API key
	tgAPIKey := os.Getenv("AB_TGAPIKEY")
	if tgAPIKey == "" {
		panic("No valid Telegram API Key specified")
	}
	//Telegram channel number
	tgChannel, _ := strconv.ParseInt(os.Getenv("AB_TGCHANNEL"), 10, 64)
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
	sbAPIKey := os.Getenv("AB_SBAPIKEY")
	if sbAPIKey == "" {
		panic("No valid Systembolaget API Key specified")
	}

	var ctx = context.Background()

	// Redis initialization
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	s := Response{}

	u, err := url.Parse("https://api-extern.systembolaget.se/sb-api-ecommerce/v1/stockbalance/store")
	if err != nil {
		panic(err)
	}

	// Create query
	q := u.Query()
	// Set systembolaget product id
	q.Set("ProductID", "508393")
	// Set systembolaget store id
	q.Set("StoreID", "0611")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Origin", "https://www.systembolaget.se")
	req.Header.Add("Authority", "api-extern.systembolaget.se")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36")
	req.Header.Add("Ocp-Apim-Subscription-Key", sbAPIKey)

	res, err := http.DefaultClient.Do(req)
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

	// Future use, support for multiple sites
	for _, site := range s {

		// Get item from redis
		val, err := rdb.Get(ctx, site.StoreID).Result()
		if err != nil {
			// If does not exist, create
			err := rdb.Set(ctx, site.StoreID, site.Stock, 0).Err()
			// If create dosent work, panic
			if err != nil {
				panic(err)
			}
		}

		// Check if stock is unchanged
		if val == strconv.Itoa(site.Stock) {
			// Debug, fix later
			//fmt.Println("No stock update, currently at " + strconv.Itoa(site.Stock))
			// else update stock and prepare message
		} else {
			err := rdb.Set(ctx, site.StoreID, site.Stock, 0).Err()
			if err != nil {
				panic(err)
			}
			message = message + "Currently left in stock: " + strconv.Itoa(site.Stock) + "\n"
			dataUpdate = true
		}

	}
	if dataUpdate == true {
		// Send message to Telegram
		t.SendM(message)
		// Debug, fix later
		//fmt.Println(message)
	}
}
