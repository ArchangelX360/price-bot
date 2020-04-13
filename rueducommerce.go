package main

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"strconv"
	"time"
)

type RueDuCommerce struct {
	productUrl string
	minPrice   float64
}

func (ru *RueDuCommerce) Name() string {
	return "RUE_DU_COMMERCE"
}

func (ru *RueDuCommerce) MinPrice() float64 {
	return ru.minPrice
}

func (ru *RueDuCommerce) IsAvailable(_ context.Context) (bool, error) {
	return true, nil // TODO
}

func (ru *RueDuCommerce) FetchPrice(ctx context.Context) (float64, error) {
	p, err := ru.getPrice(ctx)
	if err != nil {
		return InvalidPrice, fmt.Errorf("could not fetch price, got error %v", err)
	}
	return p, nil
}

func (ru *RueDuCommerce) getPrice(ctx context.Context) (float64, error) {
	var price string
	var ok bool
	err := chromedp.Run(ctx,
		chromedp.Navigate(ru.productUrl),
		chromedp.Sleep(2*time.Second),
		chromedp.AttributeValue("[itemprop='price']", "content", &price, &ok))
	if err != nil {
		return InvalidPrice, err
	}
	if !ok {
		return InvalidPrice, fmt.Errorf("failed to retrieve attribute value")
	}
	return strconv.ParseFloat(price, 64)
}
