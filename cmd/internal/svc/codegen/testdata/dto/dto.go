/**
* Generated by odin v2.0.4.
* You can edit it as your need.
 */
package dto

import "github.com/shopspring/decimal"

type StringSliceWrapper struct {
	Value []string
}

// DroppedTarget DroppedTarget has the information for one target that was dropped during relabelling.
type DroppedTarget struct {
	DiscoveredLabels map[string]StringSliceWrapper `json:"discoveredLabels"`
}

// Target Target has the information for one target.
type Target struct {
	DiscoveredLabels map[string]StringSliceWrapper `json:"discoveredLabels"`

	GlobalURL *string `json:"globalURL"`

	Health *TargetHealth `json:"health"`

	Labels *Labels `json:"labels"`

	LastError *string `json:"lastError"`

	LastScrape *string `json:"lastScrape"`

	LastScrapeDuration *float64 `json:"lastScrapeDuration"`

	ScrapePool *string `json:"scrapePool"`

	ScrapeURL *string `json:"scrapeURL"`
}

type Laptop struct {
	Price decimal.Decimal
}
