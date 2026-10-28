// Package coworkingview is the Go client for the CoworkingView catalogue API:
// coworking spaces, private offices, flexible desks and meeting rooms, with
// published prices and market rates by city.
//
// [CoworkingView] compares flexible workspace across cities such as
// [Berlin], [Madrid], [Barcelona] and [London]. This package reads the same
// catalogue over its public /v1 API at https://api.coworkingview.com.
//
// Reads are free and need no API key, no sign-up and no sales call. The API is
// rate-limited; production is the sandbox.
//
// # Quick start
//
//	client := coworkingview.NewClient()
//	page, err := client.ListProperties(ctx, &coworkingview.ListPropertiesParams{
//		City:              "berlin",
//		PricingOptionType: coworkingview.PrivateOffice,
//		PageSize:          5,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, p := range page.Data {
//		fmt.Println(p.Title, p.FromPrice)
//	}
//
// # What the API answers
//
//   - Spaces: [Client.ListProperties], [Client.AllProperties] and
//     [Client.GetProperty] (the full record, with every pricing option).
//   - Places and brands: [Client.Bootstrap] lists every country, city and
//     district; [Client.GetPlace] and [Client.GetOperator] return their profiles.
//   - Market data: [Client.MarketRates] returns the medians behind the
//     [Coworking Price Index], per city and workspace type.
//   - Search: [Client.Suggest] autocompletes cities, districts, brands and spaces.
//
// # Reading the numbers
//
// Market rates are medians over the priced spaces in a city. Every figure
// carries a SampleSize: read it before calling a figure "the market rate".
// Prices are what operators publish; [Property.WebURL] links to the live
// listing, and the [Review Index] ranks the same cities by their public reviews.
//
// # Other ways in
//
// The same catalogue is available as an MCP server for AI assistants, a CLI
// and an OpenAPI 3.1 document: see [CoworkingView for developers] and the
// [OpenAPI specification].
//
// [CoworkingView]: https://coworkingview.com
// [Berlin]: https://coworkingview.com/en/coworking/berlin
// [Madrid]: https://coworkingview.com/en/coworking/madrid
// [Barcelona]: https://coworkingview.com/en/coworking/barcelona
// [London]: https://coworkingview.com/en/coworking/london
// [Coworking Price Index]: https://coworkingview.com/en/tools/price-index
// [Review Index]: https://coworkingview.com/en/tools/review-index
// [CoworkingView for developers]: https://coworkingview.com/en/mcp
// [OpenAPI specification]: https://coworkingview.com/openapi.json
package coworkingview
