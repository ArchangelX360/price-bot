package main

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"strconv"
	"strings"
)

const (
	InStock = "InStock"
)

type MediaMarkt struct {
	productUrl string
	minPrice   float64
}

func (m *MediaMarkt) Name() string {
	return "MEDIAMARKT"
}

func (m *MediaMarkt) MinPrice() float64 {
	return m.minPrice
}

func (m *MediaMarkt) IsAvailable(ctx context.Context) (bool, error) {
	var availibility string
	var ok bool
	err := chromedp.Run(ctx,
		chromedp.Navigate(m.productUrl),
		chromedp.AttributeValue("[itemprop='availability']", "content", &availibility, &ok))
	if err != nil {
		return false, err
	}
	if !ok {
		return false, fmt.Errorf("failed to retrieve attribute value")
	}
	return strings.Compare(availibility, InStock) == 0, nil
}

func (m *MediaMarkt) FetchPrice(ctx context.Context) (float64, error) {
	price, err := m.getPrice(ctx)
	if err != nil {
		return InvalidPrice, fmt.Errorf("could not fetch price, got error %v", err)
	}
	p, err := m.convertPrice(price)
	if err != nil {
		return InvalidPrice, fmt.Errorf("could not convert price %q to number, got error %v", price, err)
	}
	return p, nil
}

func (m *MediaMarkt) convertPrice(price string) (float64, error) {
	return strconv.ParseFloat(price, 64)
}

func (m *MediaMarkt) getPrice(ctx context.Context) (string, error) {
	var price string
	var ok bool
	err := chromedp.Run(ctx,
		chromedp.Navigate(m.productUrl),
		chromedp.AttributeValue("[itemprop='price']", "content", &price, &ok))
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("failed to retrieve attribute value")
	}
	return price, nil
}
