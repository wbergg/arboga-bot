package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Response struct {
	JSON []struct {
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
}

func main() {
	requestData()
}

func requestData() {

	s := Response{}

	jsonValue := []byte(`{"productId":"508393","siteIds":["0611"]}`)

	res, err := http.Post("https://www.systembolaget.se/api/product/getstockbalance", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	json.Unmarshal(body, &s)

	fmt.Println(s.StockTextShort)
}
