package coworkingview

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// Bootstrap is every country, city and district in the catalogue, with how
// many spaces each holds. Use it to turn a name into the slug the other calls
// take.
type Bootstrap struct {
	Countries []Country `json:"countries"`
	Cities    []City    `json:"cities"`
	// Currencies maps an ISO 3166-1 country code to its local currency.
	Currencies map[string]string `json:"currencies"`
}

// Country is one country in [Bootstrap].
type Country struct {
	Slug          string `json:"slug"`
	Name          string `json:"name"`
	Code          string `json:"code"`
	PropertyCount int    `json:"propertyCount"`
}

// City is one city in [Bootstrap], with its districts.
type City struct {
	Slug          string      `json:"slug"`
	Name          string      `json:"name"`
	CountryCode   string      `json:"countryCode"`
	PropertyCount int         `json:"propertyCount"`
	Center        Coordinates `json:"center"`
	Icon          *Image      `json:"icon,omitempty"`
	Zones         []Zone      `json:"zones"`
}

// Zone is a district of a city, such as Mitte in Berlin.
type Zone struct {
	Slug          string `json:"slug"`
	Name          string `json:"name"`
	PropertyCount int    `json:"propertyCount"`
}

// Bootstrap returns every country, city and district in the catalogue.
func (c *Client) Bootstrap(ctx context.Context, locale Locale) (*Response[Bootstrap], error) {
	var out Response[Bootstrap]
	if err := c.get(ctx, "/v1/geo/bootstrap", url.Values{"locale": {c.localeOr(locale)}}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Suggestion is one autocomplete hit.
type Suggestion struct {
	// Kind is "country", "city", "zone", "operator" or "property".
	Kind          string       `json:"kind"`
	Slug          string       `json:"slug"`
	Label         string       `json:"label"`
	Sublabel      string       `json:"sublabel,omitempty"`
	Center        *Coordinates `json:"center,omitempty"`
	PropertyCount int          `json:"propertyCount,omitempty"`
}

// Suggest autocompletes a search box: q (1-100 characters) is matched against
// places, brands and spaces. limit is 1-20; 0 means the API's default of 8.
func (c *Client) Suggest(ctx context.Context, q string, limit int, locale Locale) (*Response[[]Suggestion], error) {
	if q == "" {
		return nil, fmt.Errorf("coworkingview: Suggest needs a query")
	}
	query := url.Values{"locale": {c.localeOr(locale)}, "q": {q}}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	var out Response[[]Suggestion]
	if err := c.get(ctx, "/v1/search/suggest", query, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PlaceKind is the level of a place profile.
type PlaceKind string

// Place kinds accepted by [Client.GetPlace].
const (
	KindCountry PlaceKind = "country"
	KindCity    PlaceKind = "city"
	KindZone    PlaceKind = "zone"
)

// Place is the profile of a country, city or district: its introduction,
// FAQs, how many spaces it has and their price range.
type Place struct {
	Kind             PlaceKind `json:"kind"`
	Slug             string    `json:"slug"`
	Name             string    `json:"name"`
	Intro            string    `json:"intro,omitempty"`
	ShortDescription string    `json:"shortDescription,omitempty"`
	FAQs             []FAQ     `json:"faqs"`
	PropertyCount    int       `json:"propertyCount"`
	PriceRange       *struct {
		Min Money `json:"min"`
		Max Money `json:"max"`
	} `json:"priceRange,omitempty"`
	Center *Coordinates `json:"center,omitempty"`
	Parent *PlaceRef    `json:"parent,omitempty"`
	AsOf   string       `json:"asOf"`
}

// GetPlace returns a place profile, the data behind a page such as
// [coworking in Madrid].
//
// [coworking in Madrid]: https://coworkingview.com/en/coworking/madrid
func (c *Client) GetPlace(ctx context.Context, kind PlaceKind, slug string, locale Locale) (*Response[Place], error) {
	if kind == "" || slug == "" {
		return nil, fmt.Errorf("coworkingview: GetPlace needs a kind and a slug")
	}
	var out Response[Place]
	path := "/v1/places/" + url.PathEscape(string(kind)) + "/" + url.PathEscape(slug)
	if err := c.get(ctx, path, url.Values{"locale": {c.localeOr(locale)}}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Operator is the profile of a coworking brand, such as WeWork or Regus.
type Operator struct {
	Slug          string     `json:"slug"`
	Name          string     `json:"name"`
	Logo          *Image     `json:"logo,omitempty"`
	Tagline       string     `json:"tagline,omitempty"`
	Description   string     `json:"description,omitempty"`
	FAQs          []FAQ      `json:"faqs"`
	PropertyCount int        `json:"propertyCount"`
	Cities        []PlaceRef `json:"cities"`
	// PricePosition places the brand's prices relative to each city's median.
	PricePosition *struct {
		Verdict    Verdict `json:"verdict"`
		SampleSize int     `json:"sampleSize"`
	} `json:"pricePosition,omitempty"`
	AsOf string `json:"asOf"`
}

// GetOperator returns a brand's profile, the data behind a brand page such as
// [WeWork on CoworkingView].
//
// [WeWork on CoworkingView]: https://coworkingview.com/en/brands/wework
func (c *Client) GetOperator(ctx context.Context, slug string, locale Locale) (*Response[Operator], error) {
	if slug == "" {
		return nil, fmt.Errorf("coworkingview: GetOperator needs a slug")
	}
	var out Response[Operator]
	if err := c.get(ctx, "/v1/operators/"+url.PathEscape(slug), url.Values{"locale": {c.localeOr(locale)}}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
