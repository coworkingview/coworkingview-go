package coworkingview_test

import (
	"context"
	"fmt"
	"log"

	"github.com/coworkingview/coworkingview-go"
)

// List private offices in Berlin with their entry price.
func ExampleClient_ListProperties() {
	client := coworkingview.NewClient()
	res, err := client.ListProperties(context.Background(), &coworkingview.ListPropertiesParams{
		City:              "berlin",
		PricingOptionType: coworkingview.PrivateOffice,
		PageSize:          5,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range res.Data {
		if p.FromPrice != nil {
			fmt.Printf("%s from %s\n", p.Title, p.FromPrice)
		}
	}
}

// Walk every space in a city, across pages.
func ExampleClient_AllProperties() {
	client := coworkingview.NewClient()
	count := 0
	for _, err := range client.AllProperties(context.Background(), &coworkingview.ListPropertiesParams{City: "madrid"}) {
		if err != nil {
			log.Fatal(err)
		}
		count++
	}
	fmt.Println(count, "spaces in Madrid")
}

// Read what a flex desk costs in Barcelona, with the sample behind the figure.
func ExampleClient_MarketRates() {
	client := coworkingview.NewClient()
	res, err := client.MarketRates(context.Background(), &coworkingview.MarketRatesParams{City: "barcelona"})
	if err != nil {
		log.Fatal(err)
	}
	for _, city := range res.Data.Cities {
		for _, f := range city.Figures {
			if f.OptionType == coworkingview.FlexDesk {
				fmt.Printf("%s flex desk median: %s/month (%d spaces)\n", city.City.Name, f.Median, f.PropertyCount)
			}
		}
	}
}

// Fetch one space and every product it sells.
func ExampleClient_GetProperty() {
	client := coworkingview.NewClient()
	res, err := client.GetProperty(context.Background(), "signature-potsdamer-platz-berlin", coworkingview.English)
	if coworkingview.IsNotFound(err) {
		fmt.Println("that space is not listed")
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	for _, o := range res.Data.PricingOptions {
		if o.Price != nil {
			fmt.Printf("%s: %s per %s\n", o.OptionType, o.Price, o.BillingPeriod)
		}
	}
	fmt.Println("Details:", res.Data.WebURL)
}
