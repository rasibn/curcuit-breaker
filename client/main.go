package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sony/gobreaker"
)

var cbGet *gobreaker.CircuitBreaker

func init() {
	var settings gobreaker.Settings
	settings.Name = "HTTP GET"

	settings.ReadyToTrip = func(counts gobreaker.Counts) bool {
		// if the failure rate is greater than 60% then trip the curcuit
		failureRate := float64(counts.TotalFailures) / float64(counts.Requests)
		return counts.Requests >= 10 && failureRate >= 0.6
	}

	settings.Timeout = time.Millisecond

	settings.OnStateChange = func(name string, from gobreaker.State, to gobreaker.State) {
		if to == gobreaker.StateOpen {
			log.Error().Msg("Circuit breaker is open")
		}
		if from == gobreaker.StateOpen && to == gobreaker.StateHalfOpen {
			log.Info().Msg("Going from Open to half open")
		}

		if from == gobreaker.StateHalfOpen && to == gobreaker.StateClosed {
			log.Info().Msg("Going from Half Open to Closed")
		}
	}
}

func Get(url string) ([]byte, error) {
	body, err := cbGet.Execute(func() (interface{}, error) {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("http Get request gave error")
			return nil, err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading body")
			return nil, err
		}

		return body, nil
	})
	if err != nil {
		return nil, err
	}

	return body.([]byte), nil
}

func main() {
	init()
	urlIncorrect := "http://localhost:8091"
	urlCorrect := "http://localhost:8090"

	var body []byte
	var err error

	for i := 0; i < 20; i++ {
		body, err = Get(urlIncorrect)
		if err != nil {
			log.Error().Err(err).Msg("Error")
		}
		fmt.Println(string(body))
		if i > 15 {
			urlIncorrect = urlCorrect
		}

		time.Sleep(time.Millisecond)
	}
}
