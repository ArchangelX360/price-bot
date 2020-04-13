package main

import (
	"context"
	"flag"
	"github.com/chromedp/chromedp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	debugLogger = log.New(os.Stdout, "[DEBUG  ] ", log.LstdFlags)
	infoLogger  = log.New(os.Stdout, "[INFO   ] ", log.LstdFlags)
	errorLogger = log.New(os.Stderr, "[ERROR  ] ", log.LstdFlags)
)

type Website interface {
	MinPrice() float64

	Name() string

	FetchPrice(ctx context.Context) (float64, error)

	IsAvailable(ctx context.Context) (bool, error)
}

const (
	ExpectedPrice = 330
	InvalidPrice  = 9999999
)

var (
	stmpOptionsFilepath = flag.String("smtp-filepath", "/opt/config/smtp.json", "filepath to JSON SMTP options used to send the price notifications")
	port                = flag.Int("port", 8091, "port on which Prometheus metrics (/prometheus) are exposed")
)

func init() {
	prometheus.MustRegister(lastPriceObservedGauge, lastSyncGauge, isAvailableGauge)
}

func main() {
	flag.Parse()

	notifier := NewSmtpNotifier(*stmpOptionsFilepath)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// faking user agent for usual human page to be displayed
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.163 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	go fetch(taskCtx, 30*time.Second, 0*time.Second, notifier, &MediaMarkt{
		productUrl: "https://www.mediamarkt.nl/nl/product/_nintendo-switch-rood-en-blauw-2019-revisie-1635020.html",
		minPrice:   ExpectedPrice,
	})

	go fetch(taskCtx, 30*time.Second, 1*time.Second, notifier, &Bol{
		productUrl: "https://www.bol.com/nl/p/nintendo-switch-console-met-35-eshop-tegoed-voucher-32gb-rood-blauw/9200000114613417/",
		minPrice:   ExpectedPrice,
	})

	go fetch(taskCtx, 30*time.Second, 2*time.Second, notifier, &CoolBlue{
		productUrl: "https://www.coolblue.nl/en/product/838252/nintendo-switch-2019-upgrade-red-blue.html",
		minPrice:   ExpectedPrice,
	})

	go fetch(taskCtx, 30*time.Second, 3*time.Second, notifier, &Amazon{
		productUrl: "http://amazon.nl/gp/offer-listing/B07WKNQ8JT",
		minPrice:   ExpectedPrice,
		country:    "NL",
	})

	go fetch(taskCtx, 30*time.Second, 4*time.Second, notifier, &Amazon{
		productUrl: "https://www.amazon.fr/gp/offer-listing/B07WKNQ8JT",
		minPrice:   ExpectedPrice,
		country:    "FR",
	})

	go fetch(taskCtx, 30*time.Second, 5*time.Second, notifier, &Amazon{
		productUrl: "https://www.amazon.de/gp/offer-listing/B07WKNQ8JT",
		minPrice:   ExpectedPrice,
		country:    "DE",
	})

	go fetch(taskCtx, 30*time.Second, 6*time.Second, notifier, &Cdiscount{
		productUrl: "https://www.cdiscount.com/jeux-pc-video-console/op/console-nintendo-switch-neon-nouvelle-version-me/f-103360203-45496452629.html",
		minPrice:   ExpectedPrice,
	})

	go fetch(taskCtx, 30*time.Second, 7*time.Second, notifier, &Cdiscount{
		productUrl: "https://www.rueducommerce.fr/produit/nintendo-console-switch-2019-bleue-rouge-90637297/offre-181087650",
		minPrice:   ExpectedPrice,
	})

	mux := http.NewServeMux()
	mux.Handle("/prometheus", promhttp.Handler())

	handler := cors.Default().Handler(mux)
	errorLogger.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), handler))
}

func availabilityToNumber(isAvailable bool) float64 {
	if isAvailable {
		return 1
	} else {
		return 0
	}
}

func fetch(ctx context.Context, frequency time.Duration, initialDelay time.Duration, notifier Notifier, w Website) {
	infoLogger.Printf("[%s] Starting fetching routine in %v", w.Name(), initialDelay)
	time.Sleep(initialDelay)
	for {
		newTab, closeTab := chromedp.NewContext(ctx)

		infoLogger.Printf("[%s] Checking price...", w.Name())
		price, errPrice := w.FetchPrice(newTab)
		if errPrice != nil {
			errorLogger.Printf("[%s] %v", w.Name(), errPrice)
		} else {
			lastPriceObservedGauge.With(prometheus.Labels{"website": w.Name()}).Set(price)
		}

		infoLogger.Printf("[%s] Checking availability...", w.Name())
		isAvailable, errAvailable := w.IsAvailable(newTab)
		if errAvailable != nil {
			errorLogger.Printf("[%s] %v", w.Name(), errAvailable)
		} else {
			isAvailableGauge.With(prometheus.Labels{"website": w.Name()}).Set(availabilityToNumber(isAvailable))
		}

		if errPrice == nil && errAvailable == nil {
			lastSyncGauge.With(prometheus.Labels{"website": w.Name()}).SetToCurrentTime()

			if price <= w.MinPrice() && isAvailable {
				notifier.Notify(w.Name(), price)
			} else {
				debugLogger.Printf("[%s] The product is *not* available OR too expensive (%.2f)\n", w.Name(), price)
			}
		}
		closeTab()
		time.Sleep(frequency)
	}
}
