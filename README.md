<p align="center">
  <img src="https://coworkingview.com/brand/coworkingview-wordmark.svg" alt="CoworkingView" width="320">
</p>

# coworkingview-go

[![Go Reference](https://pkg.go.dev/badge/github.com/coworkingview/coworkingview-go.svg)](https://pkg.go.dev/github.com/coworkingview/coworkingview-go)

The Go client for the [CoworkingView](https://coworkingview.com) catalogue API: coworking spaces,
private offices, flexible desks and meeting rooms, with published prices and market rates by city.

**Reads need no API key and no sign-up.** The client calls the public API at
`https://api.coworkingview.com/v1`, which is rate-limited. No dependencies beyond the standard library.

```bash
go get github.com/coworkingview/coworkingview-go
```

```go
client := coworkingview.NewClient()

res, err := client.ListProperties(ctx, &coworkingview.ListPropertiesParams{
	City:              "berlin",
	PricingOptionType: coworkingview.PrivateOffice,
	PageSize:          5,
})
if err != nil {
	log.Fatal(err)
}
for _, p := range res.Data {
	fmt.Println(p.Title, p.FromPrice)
}
```

## What you can read

| Method | Returns |
| --- | --- |
| `ListProperties`, `AllProperties` | Spaces matching a city, district, brand, workspace type, amenities, capacity, price or distance |
| `GetProperty` | One space: address, description, amenities and every pricing option |
| `Facets` | The filters available in a city, with counts |
| `Bootstrap` | Every country, city and district, with how many spaces each has |
| `Suggest` | Autocomplete over places, brands and spaces |
| `GetPlace`, `GetOperator` | Profiles of a city or district, and of a brand such as WeWork or Regus |
| `MarketRates` | Median prices per city and workspace type, the data behind the [Coworking Price Index](https://coworkingview.com/en/tools/price-index) |

`AllProperties` is an iterator (Go 1.23+) that pages for you:

```go
for p, err := range client.AllProperties(ctx, &coworkingview.ListPropertiesParams{City: "madrid"}) {
	if err != nil {
		return err
	}
	fmt.Println(p.Title)
}
```

Errors from the API are `*coworkingview.APIError` (RFC 9457 problem details). `coworkingview.IsNotFound(err)`
is true for an unknown slug and for a space that has closed.

## Reading the numbers

- Market rates are **medians over the priced spaces in a city**. Check `SampleSize` before calling a
  figure "the market rate".
- Prices are what operators publish. `Property.WebURL` links to the space's page on coworkingview.com
  with the current offer.
- The market figures are published under CC BY 4.0: credit CoworkingView and link to the source
  (`Response.Meta.Citation` has the text).

## Options

```go
coworkingview.NewClient(
	coworkingview.WithLocale(coworkingview.Spanish),       // en (default), de, es
	coworkingview.WithUserAgent("acme-office-finder/2.1"), // identify your integration
	coworkingview.WithHTTPClient(myHTTPClient),            // default: 15 s timeout
	coworkingview.WithBaseURL("http://localhost:3000"),
)
```

## More ways in

- Browse the catalogue: [coworking in Berlin](https://coworkingview.com/en/coworking/berlin),
  [Madrid](https://coworkingview.com/en/coworking/madrid), [Barcelona](https://coworkingview.com/en/coworking/barcelona),
  [London](https://coworkingview.com/en/coworking/london)
- MCP server for AI assistants, CLI and API docs: [coworkingview.com/en/mcp](https://coworkingview.com/en/mcp)
- OpenAPI 3.1: [coworkingview.com/openapi.json](https://coworkingview.com/openapi.json)
- Node-RED: [node-red-contrib-coworkingview](https://github.com/coworkingview/node-red-contrib-coworkingview)

## License

MIT © [CoworkingView](https://coworkingview.com)
