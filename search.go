package performance

import (
	"fmt"
	"net/http"
)

type SearchResult struct {
	Items []SearchItem `json:"items"`
}

type SearchItem struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Exch     string `json:"exch"`
	Type     string `json:"type"`
	ExchDisp string `json:"exchDisp"`
	TypeDisp string `json:"typeDisp"`
}

const searchHost = "https://de.finance.yahoo.com/_finance_doubledown/api/resource/searchassist;searchTerm="

var searchParams = []string{
	// "bkt=finance-DE-de-DE-def",
	"device=desktop",
	// "intl=de",
	"lang=de-DE",
	"partner=none",
	// "prid=dm010stdk9fv0",
	"region=DE",
	// "site=finance",
	"tz=Europe/Berlin",
	// "ver=0.102.1515",
	"returnMeta=false",
	// fmt.Sprintf("feature=%s", strings.Join(features, ",")),
}

var features = []string{
	"canvassOffnet",
	"ccOnMute",
	"enablePromoImage",
	"enforceFinCSP",
	"newContentAttribution",
	"relatedVideoFeature",
	"videoNativePlaylist",
	"enablePrivacyUpdate",
	"enableGuceJs",
	"enableGuceJsOverlay",
	"enableCMP",
	"enableConsentData",
	"tdCuratedWatchlists",
	"useVideoManagedPlayer",
	"enableSingleRail",
}

// Search asks Yahoo for results with the specified search value
func Search(term string) (SearchResult, error) {
	url := fmt.Sprintf("%s%s", searchHost, term)

	response, err := http.Get(url)
	if err != nil {
		return SearchResult{}, err
	}

	searchResult := SearchResult{}
	err = readBody(response.Body, &searchResult)
	if err != nil {
		return SearchResult{}, err
	}

	return searchResult, nil
}
