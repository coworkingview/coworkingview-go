package coworkingview

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
)

// The fixtures in testdata are the API contract's own published examples, so
// a decoding test here fails when the client and the contract disagree.
var fixtures = map[string]string{
	"/v1/health":                     "health",
	"/v1/properties":                 "listProperties",
	"/v1/properties/example-space":   "propertyDetail",
	"/v1/facets":                     "facets",
	"/v1/geo/bootstrap":              "geoBootstrap",
	"/v1/search/suggest":             "searchSuggest",
	"/v1/market/rates":               "marketRates",
	"/v1/operators/example-operator": "operatorProfile",
	"/v1/places/city/berlin":         "placeProfile",
}

type recorder struct {
	mu   sync.Mutex
	reqs []*http.Request
}

func (r *recorder) last() *http.Request {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reqs[len(r.reqs)-1]
}

func newTestClient(t *testing.T, opts ...Option) (*Client, *recorder) {
	t.Helper()
	rec := &recorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.mu.Lock()
		rec.reqs = append(rec.reqs, r)
		rec.mu.Unlock()
		name, ok := fixtures[r.URL.Path]
		if !ok {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"type":"https://coworkingview.com/problems/not-found","title":"Not Found","status":404,"code":"NOT_FOUND","detail":"No such space"}`))
			return
		}
		body, err := os.ReadFile("testdata/" + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return NewClient(append([]Option{WithBaseURL(srv.URL)}, opts...)...), rec
}

func TestListPropertiesQuery(t *testing.T) {
	c, rec := newTestClient(t)
	res, err := c.ListProperties(context.Background(), &ListPropertiesParams{
		City:              "berlin",
		PricingOptionType: PrivateOffice,
		Amenities:         []string{"wifi", "parking"},
		MinCapacity:       4,
		MaxPrice:          900.5,
		Near:              &Coordinates{Lat: 52.52, Lng: 13.4},
		SortBy:            SortByDistance,
		PageSize:          2,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := rec.last().URL.Query()
	want := url.Values{
		"locale": {"en"}, "city": {"berlin"}, "pricingOptionType": {"private_office"},
		"amenities": {"wifi,parking"}, "minCapacity": {"4"}, "maxPrice": {"900.5"},
		"lat": {"52.52"}, "lng": {"13.4"}, "sortBy": {"distance"}, "pageSize": {"2"},
	}
	if got.Encode() != want.Encode() {
		t.Errorf("query\n got %s\nwant %s", got.Encode(), want.Encode())
	}
	if len(res.Data) == 0 || res.Data[0].Slug == "" || res.Data[0].FromPrice == nil {
		t.Fatalf("decoded nothing useful: %+v", res.Data)
	}
	if res.Meta.Pagination == nil || res.Meta.ContractVersion == "" {
		t.Errorf("meta not decoded: %+v", res.Meta)
	}
	if ua := rec.last().Header.Get("User-Agent"); !strings.HasPrefix(ua, "coworkingview-go/") {
		t.Errorf("User-Agent = %q", ua)
	}
}

func TestNilParamsAndLocaleOption(t *testing.T) {
	c, rec := newTestClient(t, WithLocale(German))
	if _, err := c.ListProperties(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if q := rec.last().URL.RawQuery; q != "locale=de" {
		t.Errorf("query = %q, want locale=de", q)
	}
}

func TestGetProperty(t *testing.T) {
	c, _ := newTestClient(t)
	res, err := c.GetProperty(context.Background(), "example-space", "")
	if err != nil {
		t.Fatal(err)
	}
	p := res.Data
	if p.Title == "" || p.WebURL == "" || p.Address == "" || len(p.PricingOptions) == 0 {
		t.Fatalf("incomplete property: %+v", p)
	}
	for _, o := range p.PricingOptions {
		if o.Price == nil && !o.PriceOnRequest {
			t.Errorf("%s has neither a price nor priceOnRequest", o.OptionType)
		}
	}
}

func TestNotFound(t *testing.T) {
	c, _ := newTestClient(t)
	_, err := c.GetProperty(context.Background(), "gone", English)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 || apiErr.Code != "NOT_FOUND" {
		t.Fatalf("err = %v", err)
	}
	if !IsNotFound(err) {
		t.Error("IsNotFound = false")
	}
	if !strings.Contains(err.Error(), "No such space") {
		t.Errorf("message lost the detail: %v", err)
	}
}

func TestEveryEndpointDecodes(t *testing.T) {
	c, rec := newTestClient(t)
	ctx := context.Background()

	if r, err := c.Health(ctx); err != nil || r.Data.Status != "ok" {
		t.Errorf("Health: %v %+v", err, r)
	}
	if r, err := c.Bootstrap(ctx, ""); err != nil || len(r.Data.Cities) == 0 || len(r.Data.Countries) == 0 {
		t.Errorf("Bootstrap: %v %+v", err, r)
	}
	if r, err := c.Suggest(ctx, "berl", 5, ""); err != nil || len(r.Data) == 0 {
		t.Errorf("Suggest: %v %+v", err, r)
	} else if q := rec.last().URL.Query(); q.Get("q") != "berl" || q.Get("limit") != "5" {
		t.Errorf("Suggest query: %v", q)
	}
	if r, err := c.Facets(ctx, &ListPropertiesParams{City: "berlin", PageSize: 9}); err != nil || len(r.Data.Facets) == 0 {
		t.Errorf("Facets: %v %+v", err, r)
	} else if q := rec.last().URL.Query(); q.Has("pageSize") || q.Get("city") != "berlin" {
		t.Errorf("Facets sent list-only params: %v", q)
	}
	if r, err := c.MarketRates(ctx, &MarketRatesParams{City: "berlin", Currency: "USD"}); err != nil || len(r.Data.Cities) == 0 || r.Data.Cities[0].SampleSize == 0 {
		t.Errorf("MarketRates: %v %+v", err, r)
	} else if q := rec.last().URL.Query(); q.Get("currency") != "USD" {
		t.Errorf("MarketRates query: %v", q)
	}
	if r, err := c.GetOperator(ctx, "example-operator", ""); err != nil || r.Data.Name == "" {
		t.Errorf("GetOperator: %v %+v", err, r)
	}
	if r, err := c.GetPlace(ctx, KindCity, "berlin", ""); err != nil || r.Data.Name == "" {
		t.Errorf("GetPlace: %v %+v", err, r)
	}
}

func TestMissingArgumentsNeverCallTheAPI(t *testing.T) {
	c, rec := newTestClient(t)
	ctx := context.Background()
	checks := []error{}
	_, err := c.GetProperty(ctx, "", "")
	checks = append(checks, err)
	_, err = c.Suggest(ctx, "", 0, "")
	checks = append(checks, err)
	_, err = c.GetOperator(ctx, "", "")
	checks = append(checks, err)
	_, err = c.GetPlace(ctx, "", "berlin", "")
	checks = append(checks, err)
	for i, err := range checks {
		if err == nil {
			t.Errorf("check %d: no error", i)
		}
	}
	if len(rec.reqs) != 0 {
		t.Errorf("made %d requests", len(rec.reqs))
	}
}

func TestAllPropertiesPagesAndKeepsTheSeed(t *testing.T) {
	var queries []url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query())
		page := r.URL.Query().Get("page")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"slug":"space-` + page + `","title":"S","featured":false,"images":[],"amenities":[]}],` +
			`"meta":{"contractVersion":"1","seed":42,"pagination":{"page":` + page + `,"pageSize":1,"total":3,"pageCount":3}}}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	var slugs []string
	for p, err := range c.AllProperties(context.Background(), &ListPropertiesParams{City: "berlin"}) {
		if err != nil {
			t.Fatal(err)
		}
		slugs = append(slugs, p.Slug)
	}
	if strings.Join(slugs, ",") != "space-1,space-2,space-3" {
		t.Fatalf("slugs = %v", slugs)
	}
	if queries[0].Has("seed") || queries[1].Get("seed") != "42" || queries[2].Get("seed") != "42" {
		t.Errorf("seed not carried: %v", queries)
	}
	if queries[0].Get("pageSize") != "50" {
		t.Errorf("default pageSize = %s, want the API maximum 50", queries[0].Get("pageSize"))
	}
}

func TestMoneyString(t *testing.T) {
	for _, tc := range []struct {
		m    Money
		want string
	}{{Money{290, "EUR"}, "EUR 290"}, {Money{12.5, "GBP"}, "GBP 12.50"}} {
		if got := tc.m.String(); got != tc.want {
			t.Errorf("%v.String() = %q, want %q", tc.m, got, tc.want)
		}
	}
}
