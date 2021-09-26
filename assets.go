package main

import (
	"math/big"
	"net/http"
	"time"
)

type Asset struct {
	ID                string
	Rank              string
	Symbol            string
	Name              string
	Supply            *big.Float
	MaxSupply         *big.Float
	MarketCapUSD      *big.Float
	VolumeUSD24Hr     *big.Float
	PriceUSD          *big.Float
	ChangePercent24Hr *big.Float
	VWAP24Hr          *big.Float
}

func Assets(client *http.Client) (result []Asset, rateLimitReset time.Time, err error) {
	var response struct {
		Data []Asset
	}
	rateLimitReset, err = do(client, "/assets?limit=100", &response)
	result = response.Data
	return
}
