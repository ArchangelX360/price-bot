package main

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
)

type CoolBlue struct {
	productUrl string
	minPrice   float64
}

func (cb *CoolBlue) Name() string {
	return "COOLBLUE"
}

func (cb *CoolBlue) MinPrice() float64 {
	return cb.minPrice
}

func (cb *CoolBlue) IsAvailable(ctx context.Context) (bool, error) {
	var isAvailable bool
	err := chromedp.Run(ctx,
		chromedp.Navigate(cb.productUrl),
		chromedp.Evaluate(cb.isAvailableJS(), &isAvailable))
	return isAvailable, err
}

func (cb *CoolBlue) FetchPrice(ctx context.Context) (float64, error) {
	p, err := cb.getPrice(ctx)
	if err != nil {
		return InvalidPrice, fmt.Errorf("could not fetch price, got error %v", err)
	}
	return p, nil
}

func (cb *CoolBlue) getPrice(ctx context.Context) (float64, error) {
	var price float64
	err := chromedp.Run(ctx,
		chromedp.Navigate(cb.productUrl),
		chromedp.Evaluate(cb.getPriceJS(), &price))
	return price, err
}

func (cb *CoolBlue) getPriceJS() string {
	return fmt.Sprintf(`
function convertPriceStringToFloat(price) {
    return parseFloat(price.replace(',', '.').replace('-', 0))
}

function getPrice() {
    const priceElement = document.getElementsByClassName("sales-price__current")[0]
    return priceElement && priceElement.innerText ? parseFloat(priceElement.innerText) : %d;
}

getPrice();
`, InvalidPrice)
}

func (cb *CoolBlue) isAvailableJS() string {
	// return true if at least one of the product variant is available
	return `
function isAvailable() {
	const unavailable = document.querySelector('.text-color--available')
    return unavailable !== undefined && unavailable !== null
}

isAvailable();
`
}
