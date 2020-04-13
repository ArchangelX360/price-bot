package main

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"time"
)

type RueDuCommerce struct {
	productUrl string
	minPrice   float64
}

func (ru *Rueducommerce) Name() string {
	return "RUE_DU_COMMERCE"
}

func (ru *Rueducommerce) MinPrice() float64 {
	return ru.minPrice
}

func (ru *Rueducommerce) FetchPrice(ctx context.Context) (float64, error) {
	p, err := ru.getPrice(ctx)
	if err != nil {
		return InvalidPrice, fmt.Errorf("could not fetch price, got error %v", err)
	}
	return p, nil
}

func (ru *Rueducommerce) getPrice(ctx context.Context) (float64, error) {
	var price float64
	err := chromedp.Run(ctx,
		chromedp.Navigate(ru.productUrl),
		chromedp.Sleep(2*time.Second),
		chromedp.Evaluate(ru.getPriceJS(), &price))
	return price, err
}

func (ru *Rueducommerce) getPriceJS() string {
	return fmt.Sprintf(`
function convertPriceStringToFloat(price) {
    return parseFloat(price.replace('€', '.'))
}

function getPrice() {
    const priceElement = document.getElementsByClassName("price-pricesup")[0]
    return priceElement && priceElement.innerText ? convertPriceStringToFloat(priceElement.innerText) : %d;
}

getPrice();
`, InvalidPrice)
}
