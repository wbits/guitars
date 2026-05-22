// Command crawler searches external marketplaces for guitars in the collection
// and uploads price observations to the GuitarCollection API.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/wbits/guitars/internal/marketcrawler"
	"github.com/wbits/guitars/internal/marketcrawler/sources"
)

func main() {
	apiURL := flag.String("api-url", envOrDefault("GUITARS_API_URL", "http://127.0.0.1:3000"), "GuitarCollection API base URL")
	token := flag.String("token", os.Getenv("GUITARS_API_TOKEN"), "Bearer token (Cognito access token or local-dev-token)")
	guitarID := flag.String("guitar-id", "", "Crawl a single guitar id (default: all guitars)")
	skipReverb := flag.Bool("skip-reverb", false, "Skip Reverb source")
	skipEbay := flag.Bool("skip-ebay", false, "Skip eBay source")
	skipMarktplaats := flag.Bool("skip-marktplaats", false, "Skip Marktplaats source")
	flag.Parse()

	logger := log.New(os.Stdout, "crawler: ", log.LstdFlags|log.Lmsgprefix)
	client := marketcrawler.NewAPIClient(*apiURL, *token)
	if client.Token == "" {
		logger.Fatal("missing API token: pass -token or set GUITARS_API_TOKEN")
	}

	var srcs []marketcrawler.Source
	if !*skipReverb {
		srcs = append(srcs, &sources.Reverb{})
	}
	if !*skipEbay {
		srcs = append(srcs, sources.NewEbayFromEnv())
	}
	if !*skipMarktplaats {
		srcs = append(srcs, &sources.Marktplaats{})
	}
	if len(srcs) == 0 {
		logger.Fatal("no sources enabled")
	}

	runner := &marketcrawler.Runner{
		API:     client,
		Sources: srcs,
		Logger:  logger,
	}

	ctx := context.Background()
	if strings.TrimSpace(*guitarID) != "" {
		guitars, err := client.ListGuitars(ctx)
		if err != nil {
			logger.Fatalf("list guitars: %v", err)
		}
		var target marketcrawler.GuitarSummary
		found := false
		for _, g := range guitars {
			if g.ID == *guitarID {
				target = marketcrawler.GuitarSummary{
					ID:        g.ID,
					Brand:     g.Brand,
					TypeName:  g.TypeName,
					BuildYear: g.BuildYear,
				}
				found = true
				break
			}
		}
		if !found {
			logger.Fatalf("guitar %s not found", *guitarID)
		}
		if err := runner.RunGuitar(ctx, target); err != nil {
			logger.Fatalf("crawl guitar: %v", err)
		}
		return
	}
	if err := runner.RunAll(ctx); err != nil {
		logger.Fatalf("crawl all: %v", err)
	}
}

func envOrDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
