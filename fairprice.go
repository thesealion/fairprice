package fairprice

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type Ticker string

const (
	BTCUSDTicker Ticker = "BTC_USD"
	Delay               = 2 * time.Second
	StaleAfter          = 10 * time.Second
)

type TickerPrice struct {
	Ticker Ticker
	Time   time.Time
	Price  string // decimal value. example: "0", "10", "12.2", "13.2345122"
}

type PriceStreamSubscriber interface {
	SubscribePriceStream(Ticker) (chan TickerPrice, chan error)
}

type sourcePrice struct {
	Name  string
	Price TickerPrice
}

type fairPrice struct {
	streams map[Ticker]chan sourcePrice
	mutex   sync.Mutex
}

func NewFairPrice() *fairPrice {
	return &fairPrice{make(map[Ticker]chan sourcePrice), sync.Mutex{}}
}

func (fp *fairPrice) getStream(t Ticker) chan sourcePrice {
	fp.mutex.Lock()
	defer fp.mutex.Unlock()
	if _, ok := fp.streams[t]; !ok {
		fp.streams[t] = make(chan sourcePrice)
	}
	return fp.streams[t]
}

func (fp *fairPrice) AddSource(t Ticker, name string, pss PriceStreamSubscriber) {
	prChan, erChan := pss.SubscribePriceStream(t)
	stream := fp.getStream(t)
	go func() {
		for {
			select {
			case price, ok := <-prChan:
				if !ok {
					return
				}
				stream <- sourcePrice{name, price}
			case err := <-erChan:
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}()
}

func (fp *fairPrice) SubscribePriceStream(t Ticker) (chan TickerPrice, chan error) {
	priceChan := make(chan TickerPrice)
	errChan := make(chan error)
	stream := fp.getStream(t)

	go func() {
		sourcePrices := make(map[string]TickerPrice)

		timer := time.After(Delay)
		for {
			select {
			case <-timer:
				if len(sourcePrices) == 0 {
					errChan <- errors.New("no sources")
				} else {
					s := decimal.Zero
					num := decimal.Zero
					for name, price := range sourcePrices {
						pr, err := decimal.NewFromString(price.Price)
						if err != nil {
							errChan <- fmt.Errorf("invalid price for %s", name)
						} else {
							since := time.Since(price.Time)
							if since < StaleAfter {
								weight := decimal.NewFromInt(int64((StaleAfter - since) / time.Second))
								s = s.Add(pr.Mul(weight))
								num = num.Add(weight)
							}
						}
					}
					if num.Equal(decimal.Zero) {
						errChan <- errors.New("no valid sources")
					} else {
						priceChan <- TickerPrice{
							Ticker: t,
							Time:   time.Now(),
							Price:  s.Div(num).Round(2).String(),
						}
					}
				}
				timer = time.After(Delay)
			case sp := <-stream:
				sourcePrices[sp.Name] = sp.Price
			}
		}
	}()

	return priceChan, errChan
}
