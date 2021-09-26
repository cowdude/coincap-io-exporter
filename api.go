package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	responseStatus = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_response_status",
	}, []string{"uri", "code"})

	requestDuration = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_request_duration_s",
	}, []string{"uri"})
)

func parseRateLimitHeaders(headers http.Header) (reset time.Time, err error) {
	values := headers["X-Ratelimit-Reset"]
	if len(values) != 0 {
		var ts int64
		ts, err = strconv.ParseInt(values[0], 10, 64)
		if err != nil {
			err = fmt.Errorf("invalid X-Ratelimit-Reset value: %v", values[0])
			return
		}
		reset = time.Unix(ts, 0)
	}
	return
}

func do(client *http.Client, uri string, response interface{}) (rateLimitReset time.Time, err error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.coincap.io/v2"+uri, nil)
	if err != nil {
		return
	}
	if apiKey != "" {
		req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", apiKey)}
	}

	epoch := time.Now()
	res, err := client.Do(req)
	elapsed := time.Since(epoch)
	requestDuration.WithLabelValues(uri).Set(elapsed.Seconds())
	if err != nil {
		return
	}
	responseStatus.WithLabelValues(uri, fmt.Sprintf("%d", res.StatusCode)).Inc()

	rateLimitReset, err = parseRateLimitHeaders(res.Header)
	if err != nil {
		return
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	if res.StatusCode != 200 {
		err = fmt.Errorf("http error %d (%s): %s", res.StatusCode, res.Status, string(body))
		return
	}

	err = json.Unmarshal(body, &response)
	return
}
