package main

import (
	"log"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type marketCache struct {
	lvs []string
	set bool
}

var (
	lastMarkets  = make(map[string]assetCache)
	marketLabels = []string{
		"exchange_id",
		"base_id", "base_symbol",
		"quote_id", "quote_symbol",
	}
)

var (
	marketPriceQuote = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "market_price_quote",
		Help: "the amount of quote asset traded for one unit of base asset",
	}, marketLabels)

	marketPriceUSD = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "market_price_usd",
		Help: "quote price translated to USD",
	}, marketLabels)

	marketVolumeUSD24Hr = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "market_volume_usd_24h",
		Help: "volume transacted on this market in last 24 hours",
	}, marketLabels)

	marketPercentExchangeVolume = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "market_percents_exchange_volume",
		Help: "the amount of daily volume a single market transacts in relation to total daily volume of all markets on the exchange",
	}, marketLabels)

	marketTradesCount24Hr = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "market_trades_count_24h",
		Help: "number of trades on this market in the last 24 hours",
	}, marketLabels)
)

func updateMarket(market Market) []string {
	lvs := []string{
		market.ExchangeID,
		market.BaseID, market.BaseSymbol,
		market.QuoteID, market.QuoteSymbol,
	}
	bigFloatGauge(marketPriceQuote, lvs, market.PriceQuote)
	bigFloatGauge(marketPriceUSD, lvs, market.PriceUSD)
	bigFloatGauge(marketVolumeUSD24Hr, lvs, market.VolumeUSD24Hr)
	bigFloatGauge(marketPercentExchangeVolume, lvs, market.PercentExchangeVolume)
	bigFloatGauge(marketTradesCount24Hr, lvs, market.TradesCount24Hr)
	return lvs
}

func dropMarket(lvs []string) {
	marketPriceQuote.DeleteLabelValues(lvs...)
	marketPriceUSD.DeleteLabelValues(lvs...)
	marketVolumeUSD24Hr.DeleteLabelValues(lvs...)
	marketPercentExchangeVolume.DeleteLabelValues(lvs...)
	marketTradesCount24Hr.DeleteLabelValues(lvs...)
}

func updateMarkets(markets []Market) {
	for id, state := range lastMarkets {
		state.set = false
		lastMarkets[id] = state
	}
	for _, market := range markets {
		lvs := updateMarket(market)
		id := strings.Join([]string{market.ExchangeID, market.BaseID, market.QuoteID}, "/")
		set := time.Since(market.Updated.Time()) < time.Hour*4
		lastMarkets[id] = assetCache{lvs, set}
	}
	for id, state := range lastMarkets {
		if !state.set {
			log.Printf("Drop market %v", id)
			dropAsset(state.lvs)
			delete(lastMarkets, id)
		}
	}
}
