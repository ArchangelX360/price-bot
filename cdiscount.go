package main

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"strconv"
	"time"
)

type Cdiscount struct {
	productUrl string
	minPrice   float64
}

func (cd *Cdiscount) Name() string {
	return "CDISCOUNT"
}

func (cd *Cdiscount) MinPrice() float64 {
	return cd.minPrice
}

func (cd *Cdiscount) IsAvailable(_ context.Context) (bool, error) {
	return true, nil // TODO
}

func (cd *Cdiscount) FetchPrice(ctx context.Context) (float64, error) {
	p, err := cd.getPrice(ctx)
	if err != nil {
		return InvalidPrice, fmt.Errorf("could not fetch price, got error %v", err)
	}
	return p, nil
}

func (cd *Cdiscount) getPrice(ctx context.Context) (float64, error) {
	var price string
	var ok bool
	err := chromedp.Run(ctx,
		chromedp.Navigate(cd.productUrl),
		chromedp.Sleep(20*time.Second),
		chromedp.AttributeValue("[itemprop='price']", "content", &price, &ok))
	if err != nil {
		return InvalidPrice, err
	}
	if !ok {
		return InvalidPrice, fmt.Errorf("failed to retrieve attribute value")
	}
	return strconv.ParseFloat(price, 64)
}
