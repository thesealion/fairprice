package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	fp "github.com/thesealion/fairprice"
)

type mockSource struct{}

func (m mockSource) SubscribePriceStream(t fp.Ticker) (chan fp.TickerPrice, chan error) {
	priceChan := make(chan fp.TickerPrice)
	errChan := make(chan error)

	go func() {
		for {
			priceChan <- fp.TickerPrice{
				Ticker: t,
				Time:   time.Now(),
				Price:  fmt.Sprintf("%d", rand.Int()%50000),
			}
			time.Sleep(time.Duration(rand.Int()%10) * time.Second)
		}
	}()

	return priceChan, errChan
}

func main() {
	fairPrice := fp.NewFairPrice()
	for i := 0; i < 100; i++ {
		fairPrice.AddSource(fp.BTCUSDTicker, fmt.Sprintf("source%d", i), mockSource{})
	}
	pr, er := fairPrice.SubscribePriceStream(fp.BTCUSDTicker)
	for {
		select {
		case price := <-pr:
			fmt.Printf("%d, %s\n", price.Time.Unix(), price.Price)
		case err := <-er:
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
