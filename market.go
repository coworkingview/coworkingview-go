package coworkingview

import (
	"context"
	"net/url"
)

// MarketRatesParams scopes [Client.MarketRates]. Leave City and Country empty
// for every published city.
type MarketRatesParams struct {
	Locale   Locale
	City     string // city slug, e.g. "berlin"
	Country  string // country slug, e.g. "spain"
	Currency string // ISO 4217; converts every figure. Default: each city's own currency.
}

// MarketRates is the payload of [Client.MarketRates]: published city figures
// plus the cities held back for too small a sample.
type MarketRates struct {
	Cities      []CityRates       `json:"cities"`
	Unpublished []UnpublishedCity `json:"unpublished,omitempty"`
	// TotalProperties and TotalPricePoints are the sample behind every figure.
	TotalProperties  int    `json:"totalProperties"`
	TotalPricePoints int    `json:"totalPricePoints"`
	AsOf             string `json:"asOf"`
	// License, LicenseURL and Attribution say how the figures may be reused
	// (CC BY 4.0, with attribution and a link to the source).
	License     string `json:"license,omitempty"`
	LicenseURL  string `json:"licenseUrl,omitempty"`
	Attribution string `json:"attribution,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Copyright   string `json:"copyright,omitempty"`
}

// CityRates holds one city's market figures, the numbers behind its page on
// the [Coworking Price Index].
//
// [Coworking Price Index]: https://coworkingview.com/en/tools/price-index
type CityRates struct {
	City        PlaceRef `json:"city"`
	CountryCode string   `json:"countryCode,omitempty"`
	Currency    string   `json:"currency"`
	// Figures are the published medians per workspace type. Prefer them over
	// ByType, which stays for older clients and is empty when a figure would
	// have required a min or max.
	Figures []RateFigure `json:"figures,omitempty"`
	ByType  []RateByType `json:"byType"`
	// DayPass and MeetingRoom are headline medians for the two pay-as-you-go products.
	DayPass          *RateStat `json:"dayPass,omitempty"`
	MeetingRoom      *RateStat `json:"meetingRoom,omitempty"`
	PropertyCount    int       `json:"propertyCount,omitempty"`
	OperatorCount    int       `json:"operatorCount,omitempty"`
	InsufficientData bool      `json:"insufficientData,omitempty"`
	// Buildings lists each space's entry price, with its operator.
	Buildings []BuildingPrice `json:"buildings,omitempty"`
	// EntryMedian, EntryMin and EntryMax describe each space's entry price.
	EntryMedian *Money `json:"entryMedian,omitempty"`
	EntryMin    *Money `json:"entryMin,omitempty"`
	EntryMax    *Money `json:"entryMax,omitempty"`
	Min         *Money `json:"min,omitempty"`
	Max         *Money `json:"max,omitempty"`
	P25         *Money `json:"p25,omitempty"`
	P75         *Money `json:"p75,omitempty"`
	// SampleSize is the number of priced spaces in the city backing the
	// aggregate, across all workspace types.
	SampleSize int `json:"sampleSize"`
}

// RateFigure is one workspace type's market figure in a city.
type RateFigure struct {
	OptionType PricingOptionType `json:"optionType"`
	Median     Money             `json:"median"`
	// P25 and P75 are nil when the sample is too small for quartiles.
	P25           *Money `json:"p25"`
	P75           *Money `json:"p75"`
	SampleSize    int    `json:"sampleSize"`
	PropertyCount int    `json:"propertyCount"`
	OperatorCount int    `json:"operatorCount"`
	// InsufficientData is always false on a published figure.
	InsufficientData bool `json:"insufficientData"`
}

// RateByType is the legacy per-type aggregate. See [CityRates.Figures].
type RateByType struct {
	OptionType PricingOptionType `json:"optionType"`
	Median     Money             `json:"median"`
	Min        Money             `json:"min"`
	Max        Money             `json:"max"`
	// SampleSize is how many priced spaces of this type the figure comes from.
	SampleSize int `json:"sampleSize"`
}

// BuildingPrice is one space's entry price in [CityRates.Buildings].
type BuildingPrice struct {
	Price        Money  `json:"price"`
	Name         string `json:"name,omitempty"`
	Slug         string `json:"slug,omitempty"`
	OperatorName string `json:"operatorName,omitempty"`
	OperatorSlug string `json:"operatorSlug,omitempty"`
}

// RateStat is a single median with the sample behind it.
type RateStat struct {
	Median           Money `json:"median"`
	SampleSize       int   `json:"sampleSize"`
	PropertyCount    int   `json:"propertyCount,omitempty"`
	OperatorCount    int   `json:"operatorCount,omitempty"`
	InsufficientData bool  `json:"insufficientData,omitempty"`
}

// UnpublishedCity is a city with spaces but too few prices to publish a figure.
type UnpublishedCity struct {
	City             PlaceRef `json:"city"`
	InsufficientData bool     `json:"insufficientData"`
}

// MarketRates returns what flexible workspace costs, per city and workspace
// type: medians, quartiles and the sample size behind each figure.
func (c *Client) MarketRates(ctx context.Context, params *MarketRatesParams) (*Response[MarketRates], error) {
	if params == nil {
		params = &MarketRatesParams{}
	}
	q := url.Values{"locale": {c.localeOr(params.Locale)}}
	if params.City != "" {
		q.Set("city", params.City)
	}
	if params.Country != "" {
		q.Set("country", params.Country)
	}
	if params.Currency != "" {
		q.Set("currency", params.Currency)
	}
	var out Response[MarketRates]
	if err := c.get(ctx, "/v1/market/rates", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
