package main

import (
	"log"
	"math/big"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type assetCache struct {
	lvs []string
	set bool
}

var (
	lastAssets  = make(map[string]assetCache)
	assetLabels = []string{
		"id", "symbol", "name",
	}
)

var (
	assetSupply = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "asset_supply",
		Help: "available supply for trading",
	}, assetLabels)

	assetMaxSupply = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "asset_max_supply",
		Help: "total quantity of asset issued",
	}, assetLabels)

	assetMarketCapUSD = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "asset_market_cap_usd",
		Help: "supply x price",
	}, assetLabels)

	assetVolumeUSD24hr = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "asset_volume_usd_24h",
		Help: "quantity of trading volume represented in USD over the last 24 hours",
	}, assetLabels)

	assetPriceUSD = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "asset_price_usd",
		Help: "volume-weighted price based on real-time market data, translated to USD",
	}, assetLabels)

	assetChangePercents24hr = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "asset_change_percents_24h",
		Help: "the direction and value change in the last 24 hours",
	}, assetLabels)

	assetVWAP24hr = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "asset_vwap_24h",
		Help: "Volume Weighted Average Price in the last 24 hours",
	}, assetLabels)
)

func bigFloatGauge(gvec *prometheus.GaugeVec, lvs []string, val *big.Float) {
	if val == nil {
		gvec.DeleteLabelValues(lvs...)
		return
	}
	x, _ := val.Float64()
	gvec.WithLabelValues(lvs...).Set(x)
}

func updateAsset(asset Asset) []string {
	lvs := []string{asset.ID, asset.Symbol, asset.Name}
	bigFloatGauge(assetSupply, lvs, asset.Supply)
	bigFloatGauge(assetMaxSupply, lvs, asset.MaxSupply)
	bigFloatGauge(assetMarketCapUSD, lvs, asset.MarketCapUSD)
	bigFloatGauge(assetVolumeUSD24hr, lvs, asset.VolumeUSD24Hr)
	bigFloatGauge(assetPriceUSD, lvs, asset.PriceUSD)
	bigFloatGauge(assetChangePercents24hr, lvs, asset.ChangePercent24Hr)
	bigFloatGauge(assetVWAP24hr, lvs, asset.VWAP24Hr)
	return lvs
}

func dropAsset(lvs []string) {
	assetSupply.DeleteLabelValues(lvs...)
	assetMaxSupply.DeleteLabelValues(lvs...)
	assetMarketCapUSD.DeleteLabelValues(lvs...)
	assetVolumeUSD24hr.DeleteLabelValues(lvs...)
	assetPriceUSD.DeleteLabelValues(lvs...)
	assetChangePercents24hr.DeleteLabelValues(lvs...)
	assetVWAP24hr.DeleteLabelValues(lvs...)
}

func updateAssets(assets []Asset) {
	for id, state := range lastAssets {
		state.set = false
		lastAssets[id] = state
	}
	for _, asset := range assets {
		lvs := updateAsset(asset)
		lastAssets[asset.ID] = assetCache{lvs, true}
	}
	for id, state := range lastAssets {
		if !state.set {
			log.Printf("Drop asset %v", id)
			dropAsset(state.lvs)
			delete(lastAssets, id)
		}
	}
}
