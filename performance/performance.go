package performance

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aodin/date"
)

type yahooHistoricResult struct {
	Spark struct {
		Result []struct {
			Symbol   string `json:"symbol"`
			Response []struct {
				Timestamps []int64 `json:"timestamp"`
				Indicators struct {
					AdjClose []struct {
						AdjClose []float64 `json:"adjclose"`
					} `json:"adjclose"`
				} `json:"indicators"`
			} `json:"response"`
		} `json:"result"`
	} `json:"spark"`
}

type HistoricData struct {
	Symbol       string          `json:"symbol"`
	CurrentPrice float64         `json:"currentPrice"`
	Data         []HistoricPoint `json:"data"`
}

type HistoricPoint struct {
	Timestamp date.Date `json:"timestamp"`
	Price     float64   `json:"price"`
}

type TimeUnit string

//The valid time units for the range and interval request parameters
const (
	Year  TimeUnit = "y"
	Month TimeUnit = "mo"
	Day   TimeUnit = "d"
)

type HistoricParams struct {
	Range    Range
	Interval Interval
}

type Range struct {
	Amount   int
	TimeUnit TimeUnit
}

type Interval struct {
	Amount   int
	TimeUnit TimeUnit
}

// NewHistoricParams returns a new HistoricParams struct with a range of 1 year and quote interval of 1 day
func NewHistoricParams() HistoricParams {
	return HistoricParams{
		Range: Range{
			Amount:   1,
			TimeUnit: Year,
		},
		Interval: Interval{
			Amount:   1,
			TimeUnit: Day,
		},
	}
}

var currentInterval Interval

// GetHistoricData looks up the given sybols in the Yahoo Finance API and returns all the historic quotes for a specified range and interval
func GetHistoricData(params HistoricParams, symbols ...string) ([]HistoricData, error) {
	if ok := validParams(params); !ok {
		return []HistoricData{}, fmt.Errorf("Params for fetching of historical data invalid")
	}

	currentInterval = params.Interval

	scope := fmt.Sprintf("range=%d%s", params.Range.Amount, params.Range.TimeUnit)
	interval := fmt.Sprintf("interval=%d%s", params.Interval.Amount, params.Interval.TimeUnit)
	historicParams := fmt.Sprintf("%s&%s", scope, interval)
	historicURL := "https://query1.finance.yahoo.com/v7/finance/spark?symbols="
	url := fmt.Sprintf("%s%s&%s", historicURL, strings.Join(symbols, ","), historicParams)

	response, err := http.Get(url)
	if err != nil {
		return []HistoricData{}, err
	}

	yahooHistoricResult := yahooHistoricResult{}
	err = readBody(response.Body, &yahooHistoricResult)
	if err != nil {
		return []HistoricData{}, err
	}

	historicData := mapYahooHistoricDataToMyHistoricData(yahooHistoricResult)

	return historicData, nil
}

func validParams(params HistoricParams) bool {
	if params.Interval.Amount <= 0 {
		return false
	}
	if params.Range.Amount <= 0 {
		return false
	}

	return true
}

func mapYahooHistoricDataToMyHistoricData(yahoo yahooHistoricResult) []HistoricData {
	historicData := make([]HistoricData, 0)
	result := yahoo.Spark.Result

	for _, v := range result {
		quotes := v.Response[0].Indicators.AdjClose[0].AdjClose

		symbolData := HistoricData{
			Symbol:       v.Symbol,
			CurrentPrice: quotes[len(quotes)-1],
			Data:         getHistoricPoints(quotes, v.Response[0].Timestamps),
		}

		historicData = append(historicData, symbolData)
	}

	return historicData
}

func getHistoricPoints(quotes []float64, timestamps []int64) []HistoricPoint {
	historicPoints := make([]HistoricPoint, 0)

	for i := range quotes {
		unix := time.Unix(timestamps[i], 0).UTC()
		historicPoint := HistoricPoint{
			Timestamp: date.New(unix.Year(), unix.Month(), unix.Day()),
			Price:     quotes[i],
		}

		historicPoints = append(historicPoints, historicPoint)
	}

	historicPoints = fixMissingHistoricPoints(historicPoints)

	return historicPoints
}

func fixMissingHistoricPoints(points []HistoricPoint) []HistoricPoint {
	result := make([]HistoricPoint, 0)
	result = append(result, points[0])

	for i := 1; i < len(points); i++ {
		timestamp := points[i-1].Timestamp
		timestampToCheck := getTimestampToCheck(timestamp)
		monthIsMissing := timestampToCheck.Before(points[i].Timestamp)

		if monthIsMissing {
			insertMissingDataPoint(i, timestampToCheck, &result, points[i])
		}

		result = append(result, points[i])
	}

	return result
}

func insertMissingDataPoint(i int, timestamp date.Date, points *[]HistoricPoint, nextPoint HistoricPoint) {
	newDataPoint := HistoricPoint{
		Timestamp: timestamp,
		Price:     guessCorrectPrice(i, *points, nextPoint.Price),
	}
	*points = append(*points, newDataPoint)
}

func getTimestampToCheck(timestamp date.Date) date.Date {
	if currentInterval.TimeUnit == "mo" {
		month := time.Month(currentInterval.Amount)
		firstDayOfNextMonth := date.New(timestamp.Year(), timestamp.Month()+month, 1)
		lastDayOfNextMonth := firstDayOfNextMonth.AddDate(0, 1, -1)
		return lastDayOfNextMonth
	}

	if currentInterval.TimeUnit == "d" {
		return timestamp.AddDate(0, 0, 1)
	}

	return date.Today()
}

func guessCorrectPrice(i int, points []HistoricPoint, nextPrice float64) float64 {
	previousPrice := points[i-1].Price
	return (previousPrice + nextPrice) / 2
}
