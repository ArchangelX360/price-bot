package main

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"time"
)

type Bol struct {
	productUrl string
	minPrice   float64
}

func (b *Bol) Name() string {
	return "BOL"
}

func (b *Bol) MinPrice() float64 {
	return b.minPrice
}

func (b *Bol) IsAvailable(ctx context.Context) (bool, error) {
	var isAvailable bool
	err := chromedp.Run(ctx,
		chromedp.Navigate(b.productUrl),
		chromedp.Evaluate(b.isAvailableJS(), &isAvailable))
	return isAvailable, err
}

func (b *Bol) FetchPrice(ctx context.Context) (float64, error) {
	p, err := b.getPrice(ctx)
	if err != nil {
		return InvalidPrice, fmt.Errorf("could not fetch price, got error %v", err)
	}
	return p, nil
}

func (b *Bol) getPrice(ctx context.Context) (float64, error) {
	var price float64
	err := chromedp.Run(ctx,
		chromedp.Navigate(b.productUrl),
		chromedp.Sleep(2*time.Second),
		chromedp.Evaluate(b.getPriceJS(), &price))
	return price, err
}

func (b *Bol) getPriceJS() string {
	return fmt.Sprintf(`
function convertPriceStringToFloat(price) {
    return parseFloat(price.replace('\n', '.').replace('-', 0))
}

function getPrice() {
    const priceElement = document.getElementsByClassName("promo-price")[0];
    return priceElement && priceElement.innerText ? convertPriceStringToFloat(priceElement.innerText) : %d;
}

getPrice();
`, InvalidPrice)
}

func (b *Bol) isAvailableJS() string {
	// return true if at least one of the product variant is available
	return `
function isAvailable() {
	const inStockText = "Op voorraad"

	const unavailable = document.querySelector('.buy-block__highlight')
    return unavailable !== undefined && unavailable !== null && unavailable.innerText === inStockText
}

isAvailable();
`
}
