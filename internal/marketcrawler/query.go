package marketcrawler

import (
	"fmt"
	"strings"
)

// Common model families used to broaden marketplace searches when the full
// type name is too specific for Marktplaats or eBay.
var modelFamilyPhrases = []string{
	"Les Paul",
	"Stratocaster",
	"Telecaster",
	"Jazzmaster",
	"Jaguar",
	"Flying V",
	"Explorer",
	"SG",
	"Mustang",
	"Precision Bass",
	"Jazz Bass",
	"ES-335",
	"ES-339",
	"Semi-Hollow",
}

// SearchQuery builds the most specific marketplace search phrase from guitar attributes.
func SearchQuery(g GuitarSummary) string {
	queries := SearchQueries(g)
	if len(queries) == 0 {
		return ""
	}
	return queries[0]
}

// SearchQueries returns progressively broader search phrases. Sources should try
// each query until one returns results.
func SearchQueries(g GuitarSummary) []string {
	brand := strings.TrimSpace(g.Brand)
	model := strings.TrimSpace(g.TypeName)
	year := ""
	if g.BuildYear > 0 {
		year = fmt.Sprintf("%d", g.BuildYear)
	}

	seen := make(map[string]struct{})
	out := make([]string, 0, 6)
	add := func(parts ...string) {
		filtered := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				filtered = append(filtered, part)
			}
		}
		if len(filtered) == 0 {
			return
		}
		query := strings.Join(filtered, " ")
		if _, ok := seen[query]; ok {
			return
		}
		seen[query] = struct{}{}
		out = append(out, query)
	}

	if year != "" {
		add(brand, model, year)
	}
	add(brand, model)
	for _, family := range modelFamiliesIn(model) {
		if year != "" {
			add(brand, family, year)
		}
		add(brand, family)
	}
	if brand != "" {
		add(brand)
	}
	return out
}

func modelFamiliesIn(model string) []string {
	if model == "" {
		return nil
	}
	lower := strings.ToLower(model)
	out := make([]string, 0, 2)
	seen := make(map[string]struct{})
	for _, family := range modelFamilyPhrases {
		if !strings.Contains(lower, strings.ToLower(family)) {
			continue
		}
		if _, ok := seen[family]; ok {
			continue
		}
		seen[family] = struct{}{}
		out = append(out, family)
	}
	return out
}
