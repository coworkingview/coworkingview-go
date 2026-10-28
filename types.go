package coworkingview

import "fmt"

// Locale is the language of the text fields in a response.
type Locale string

// The languages the API serves.
const (
	English Locale = "en"
	German  Locale = "de"
	Spanish Locale = "es"
)

// PricingOptionType is a kind of workspace a space sells.
type PricingOptionType string

// Every pricing option type the catalogue uses.
const (
	PrivateOffice PricingOptionType = "private_office"
	ExclusiveArea PricingOptionType = "exclusive_bereich" // a whole floor or wing for one team
	FixedDesk     PricingOptionType = "fix_desk"
	FlexDesk      PricingOptionType = "flex_desk"
	DayPass       PricingOptionType = "day_pass"
	MeetingRoom   PricingOptionType = "meeting_room"
	VirtualOffice PricingOptionType = "virtual_office"
	EventSpace    PricingOptionType = "event_space"
	Membership    PricingOptionType = "membership"
)

// BillingPeriod is what a price covers.
type BillingPeriod string

// Billing periods.
const (
	PerHour  BillingPeriod = "hour"
	PerDay   BillingPeriod = "day"
	PerWeek  BillingPeriod = "week"
	PerMonth BillingPeriod = "month"
	PerYear  BillingPeriod = "year"
	Once     BillingPeriod = "once"
)

// Verdict places a price relative to its city's median.
type Verdict string

// Price verdicts.
const (
	Below  Verdict = "below"
	Around Verdict = "around"
	Above  Verdict = "above"
)

// Money is an amount in an ISO 4217 currency (EUR, GBP, USD, CHF, AED, VES).
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

func (m Money) String() string {
	if m.Amount == float64(int64(m.Amount)) {
		return fmt.Sprintf("%s %d", m.Currency, int64(m.Amount))
	}
	return fmt.Sprintf("%s %.2f", m.Currency, m.Amount)
}

// PlaceRef names a city, district or brand by slug.
type PlaceRef struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// Coordinates are WGS84 degrees.
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Image is a photo or logo with its intrinsic size.
type Image struct {
	URL       string `json:"url"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Alt       string `json:"alt"`
	ThumbHash string `json:"thumbHash,omitempty"`
}

// OperatorRef is the brand that runs a space.
type OperatorRef struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	Logo *Image `json:"logo,omitempty"`
}

// MarketPosition compares a space's entry price with its city's median for
// the same kind of workspace.
type MarketPosition struct {
	Verdict    Verdict `json:"verdict"`
	Median     Money   `json:"median"`
	SampleSize int     `json:"sampleSize"`
}

// FAQ is one question and answer from a place or brand profile.
type FAQ struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// Meta travels with every response.
type Meta struct {
	ContractVersion string      `json:"contractVersion"`
	Pagination      *Pagination `json:"pagination,omitempty"`
	Truncated       bool        `json:"truncated,omitempty"`
	// Seed is the shuffle seed of a random ordering. Pass it back in
	// ListPropertiesParams.Seed to page through the same order.
	Seed            *int      `json:"seed,omitempty"`
	Citation        *Citation `json:"citation,omitempty"`
	CatalogueNotice string    `json:"catalogueNotice,omitempty"`
}

// Pagination describes one page of a list.
type Pagination struct {
	Page      int `json:"page"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
	PageCount int `json:"pageCount"`
}

// Citation is how to credit CoworkingView when you republish a figure.
type Citation struct {
	Source   string `json:"source"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Accessed string `json:"accessed"`
	Text     string `json:"text"`
}
