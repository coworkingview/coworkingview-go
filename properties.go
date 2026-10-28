package coworkingview

import (
	"context"
	"fmt"
	"iter"
	"net/url"
	"strconv"
	"strings"
)

// SortBy orders a property list.
type SortBy string

// Orderings accepted by [Client.ListProperties].
const (
	// SortRandom is the default: a seeded shuffle, so no space is always first.
	SortRandom     SortBy = "random"
	SortPriceAsc   SortBy = "price-asc"
	SortNewest     SortBy = "date-desc"
	SortOldest     SortBy = "date-asc"
	SortByDistance SortBy = "distance" // needs Near
)

// ListPropertiesParams filters [Client.ListProperties]. Every field is
// optional; slugs come from [Client.Bootstrap] or [Client.Suggest].
type ListPropertiesParams struct {
	Locale            Locale
	Country           string // country slug, e.g. "germany"
	City              string // city slug, e.g. "berlin"
	Zone              string // district slug, e.g. "mitte"
	Operator          string // brand slug, e.g. "wework"
	PricingOptionType PricingOptionType
	Amenities         []string
	MinCapacity       int
	MinPrice          float64
	MaxPrice          float64
	Currency          string // ISO 4217; MinPrice/MaxPrice are in this currency
	SortBy            SortBy
	// Near searches around a point; RadiusMeters (up to 200 km) bounds it and
	// SortByDistance orders by it.
	Near         *Coordinates
	RadiusMeters int
	// Address narrows to listings at that street address, e.g. "Domstraße 17".
	// It is not geocoded: to search around an address, use Near.
	Address  string
	Page     int // 1-based, up to 500
	PageSize int // 1-50, default 20
	// Seed keeps a SortRandom order stable across pages: reuse Meta.Seed from
	// the first page. AllProperties does this for you.
	Seed int
}

func (p *ListPropertiesParams) query(c *Client) url.Values {
	if p == nil {
		p = &ListPropertiesParams{}
	}
	q := url.Values{"locale": {c.localeOr(p.Locale)}}
	set := func(k, v string) {
		if v != "" {
			q.Set(k, v)
		}
	}
	setInt := func(k string, v int) {
		if v != 0 {
			q.Set(k, strconv.Itoa(v))
		}
	}
	setFloat := func(k string, v float64) {
		if v != 0 {
			q.Set(k, strconv.FormatFloat(v, 'f', -1, 64))
		}
	}
	set("country", p.Country)
	set("city", p.City)
	set("zone", p.Zone)
	set("operator", p.Operator)
	set("pricingOptionType", string(p.PricingOptionType))
	set("amenities", strings.Join(p.Amenities, ","))
	setInt("minCapacity", p.MinCapacity)
	setFloat("minPrice", p.MinPrice)
	setFloat("maxPrice", p.MaxPrice)
	set("currency", p.Currency)
	set("sortBy", string(p.SortBy))
	if p.Near != nil {
		setFloat("lat", p.Near.Lat)
		setFloat("lng", p.Near.Lng)
	}
	setInt("radiusMeters", p.RadiusMeters)
	set("address", p.Address)
	setInt("page", p.Page)
	setInt("pageSize", p.PageSize)
	setInt("seed", p.Seed)
	return q
}

// PropertySummary is one space in a list.
type PropertySummary struct {
	Slug              string          `json:"slug"`
	Title             string          `json:"title"`
	CountryCode       string          `json:"countryCode,omitempty"`
	City              *PlaceRef       `json:"city,omitempty"`
	Zone              *PlaceRef       `json:"zone,omitempty"`
	Operator          *OperatorRef    `json:"operator,omitempty"`
	Coordinates       *Coordinates    `json:"coordinates,omitempty"`
	FromPrice         *Money          `json:"fromPrice,omitempty"`
	FromBillingPeriod BillingPeriod   `json:"fromBillingPeriod,omitempty"`
	Featured          bool            `json:"featured"`
	UpdatedAt         string          `json:"updatedAt,omitempty"`
	Excerpt           string          `json:"excerpt,omitempty"`
	Images            []Image         `json:"images"`
	Amenities         []string        `json:"amenities"`
	Capacity          int             `json:"capacity,omitempty"`
	MarketPosition    *MarketPosition `json:"marketPosition,omitempty"`
	// DistanceMeters is set when the list was searched near a point.
	DistanceMeters *float64 `json:"distanceMeters,omitempty"`
}

// ListProperties returns one page of spaces matching params, the same list as
// a city page such as [coworking in Berlin].
//
// [coworking in Berlin]: https://coworkingview.com/en/coworking/berlin
func (c *Client) ListProperties(ctx context.Context, params *ListPropertiesParams) (*Response[[]PropertySummary], error) {
	var out Response[[]PropertySummary]
	if err := c.get(ctx, "/v1/properties", params.query(c), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AllProperties walks every page of a list, yielding one space at a time. It
// stops at the first error, which it yields. Leave Page empty; PageSize sets
// how many spaces each request fetches.
//
//	for p, err := range client.AllProperties(ctx, &coworkingview.ListPropertiesParams{City: "madrid"}) {
//		if err != nil {
//			return err
//		}
//		fmt.Println(p.Title)
//	}
func (c *Client) AllProperties(ctx context.Context, params *ListPropertiesParams) iter.Seq2[PropertySummary, error] {
	return func(yield func(PropertySummary, error) bool) {
		p := ListPropertiesParams{}
		if params != nil {
			p = *params
		}
		if p.PageSize == 0 {
			p.PageSize = 50
		}
		for page := 1; ; page++ {
			p.Page = page
			res, err := c.ListProperties(ctx, &p)
			if err != nil {
				yield(PropertySummary{}, err)
				return
			}
			for _, item := range res.Data {
				if !yield(item, nil) {
					return
				}
			}
			// Keep the shuffle stable across pages of a random order.
			if p.Seed == 0 && res.Meta.Seed != nil {
				p.Seed = *res.Meta.Seed
			}
			if res.Meta.Pagination == nil || page >= res.Meta.Pagination.PageCount || len(res.Data) == 0 {
				return
			}
		}
	}
}

// PricingOption is one product a space sells, with its published price.
type PricingOption struct {
	OptionType PricingOptionType `json:"optionType"`
	// Title is the operator's own name for the product, e.g. "Team office for 6".
	Title string `json:"title"`
	// Price is nil when PriceOnRequest is true: the operator does not publish it.
	Price          *Money        `json:"price,omitempty"`
	PriceOnRequest bool          `json:"priceOnRequest"`
	BillingPeriod  BillingPeriod `json:"billingPeriod"`
	MinCapacity    int           `json:"minCapacity,omitempty"`
	MaxCapacity    int           `json:"maxCapacity,omitempty"`
}

// Property is the full record for one space.
type Property struct {
	Slug              string          `json:"slug"`
	Title             string          `json:"title"`
	CountryCode       string          `json:"countryCode,omitempty"`
	City              *PlaceRef       `json:"city,omitempty"`
	Zone              *PlaceRef       `json:"zone,omitempty"`
	Operator          *OperatorRef    `json:"operator,omitempty"`
	Coordinates       *Coordinates    `json:"coordinates,omitempty"`
	FromPrice         *Money          `json:"fromPrice,omitempty"`
	FromBillingPeriod BillingPeriod   `json:"fromBillingPeriod,omitempty"`
	Featured          bool            `json:"featured"`
	UpdatedAt         string          `json:"updatedAt,omitempty"`
	Description       string          `json:"description"`
	Address           string          `json:"address"`
	Images            []Image         `json:"images"`
	Amenities         []string        `json:"amenities"`
	Capacity          int             `json:"capacity,omitempty"`
	PricingOptions    []PricingOption `json:"pricingOptions"`
	MarketPosition    *MarketPosition `json:"marketPosition,omitempty"`
	// WebURL is the space's page on coworkingview.com, where the current offer
	// and an enquiry form live.
	WebURL string `json:"webUrl"`
}

// GetProperty returns one space by slug. A space that closed answers 410;
// test for both cases with [IsNotFound].
func (c *Client) GetProperty(ctx context.Context, slug string, locale Locale) (*Response[Property], error) {
	if slug == "" {
		return nil, fmt.Errorf("coworkingview: GetProperty needs a slug")
	}
	var out Response[Property]
	q := url.Values{"locale": {c.localeOr(locale)}}
	if err := c.get(ctx, "/v1/properties/"+url.PathEscape(slug), q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Facet is one filter dimension with the values present in a result set.
type Facet struct {
	Key    string       `json:"key"`
	Label  string       `json:"label"`
	Values []FacetValue `json:"values"`
}

// FacetValue is one filter value and how many spaces carry it.
type FacetValue struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

// Facets is the payload of [Client.Facets].
type Facets struct {
	Facets     []Facet `json:"facets"`
	PriceRange *struct {
		Min Money `json:"min"`
		Max Money `json:"max"`
	} `json:"priceRange,omitempty"`
	CapacityRange *struct {
		Min int `json:"min"`
		Max int `json:"max"`
	} `json:"capacityRange,omitempty"`
}

// Facets returns the filters available for a scope (Locale, Country, City,
// Zone, Operator, PricingOptionType and Currency are read from params), with
// counts: the amenities, districts and brands you can filter a city by.
func (c *Client) Facets(ctx context.Context, params *ListPropertiesParams) (*Response[Facets], error) {
	full := params.query(c)
	q := url.Values{}
	for _, k := range []string{"locale", "country", "city", "zone", "operator", "pricingOptionType", "currency"} {
		if v := full.Get(k); v != "" {
			q.Set(k, v)
		}
	}
	var out Response[Facets]
	if err := c.get(ctx, "/v1/facets", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
