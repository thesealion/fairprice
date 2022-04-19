# Fairprice

**Fairprice** is a library for combining multiple streams of ticker prices into a single price stream.

It implements PriceStreamSubscriber interface and uses upstream
PriceStreamSubscriber instances as data sources. You can see a usage example in
`cmd/fairprice/main.go`.

The idea is to continually receive price updates for all the sources keeping
the latest value in memory and recalculate the fair price every minute (as
defined in `Delay` constant). The time of the updated for each source is taken
into account. If the price is too old (configured by `StaleAfter` constant), it
is not used. The calculation uses a weighted average formula in which the weight
is defined as the number of seconds left until the price is considered stale. Thus,
older price values have smaller influence on the result.
