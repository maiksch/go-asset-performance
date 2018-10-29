package performance

import (
	"fmt"
	"net/http"
	"strings"
)

type ValueResult struct {
	Assets []ValueItem
}

func (vr ValueResult) FindBySymbol(symbol string) (ValueItem, error) {
	for _, item := range vr.Assets {
		if item.Symbol == symbol {
			return item, nil
		}
	}

	return ValueItem{}, fmt.Errorf("No item found")
}

type ValueItem struct {
	Symbol       string
	CurrentPrice float64
}

type yahooResult struct {
	QuoteResponse struct {
		Result []struct {
			Symbol             string  `json:"symbol"`
			RegularMarketPrice float64 `json:"regularMarketPrice"`
		} `json:"result"`
	} `json:"quoteResponse"`
}

const performanceURL = "https://query1.finance.yahoo.com/v7/finance/quote?symbols="

// GetValue asks the Yahoo Finance API for the current values of the given symbols
func GetValue(symbols []string) (ValueResult, error) {
	joinedSymbols := strings.Join(symbols, ",")
	url := fmt.Sprintf("%s%s", performanceURL, joinedSymbols)

	response, err := http.Get(url)
	if err != nil {
		return ValueResult{}, err
	}

	yahooResult := yahooResult{}
	err = readBody(response.Body, &yahooResult)
	if err != nil {
		return ValueResult{}, err
	}

	result, err := mapYahooResultToValueResult(yahooResult)
	if err != nil {
		return ValueResult{}, err
	}

	return result, nil
}

func mapYahooResultToValueResult(yahoo yahooResult) (ValueResult, error) {
	items := make([]ValueItem, 0)

	for _, yahooItem := range yahoo.QuoteResponse.Result {
		item := ValueItem{
			Symbol:       yahooItem.Symbol,
			CurrentPrice: yahooItem.RegularMarketPrice,
		}

		items = append(items, item)
	}

	result := ValueResult{items}
	return result, nil
}
