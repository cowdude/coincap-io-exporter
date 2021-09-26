package main

import (
	"math/big"
	"net/http"
	"time"
)

type Market struct {
	ExchangeID            string
	Rank                  string
	BaseSymbol            string
	BaseID                string
	QuoteSymbol           string
	QuoteID               string
	PriceQuote            *big.Float
	PriceUSD              *big.Float
	VolumeUSD24Hr         *big.Float
	PercentExchangeVolume *big.Float
	TradesCount24Hr       *big.Float
	Updated               Timestamp
}

func Markets(client *http.Client) (result []Market, rateLimitReset time.Time, err error) {
	var response struct {
		Data []Market
	}
	rateLimitReset, err = do(client, "/markets?limit=2000", &response)
	result = response.Data
	return
}
