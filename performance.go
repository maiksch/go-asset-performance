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

func GetHistoricData(params HistoricParams, symbols ...string) ([]HistoricData, error) {
	if ok := validParams(params); !ok {
		return []HistoricData{}, fmt.Errorf("Params for fetching of historical data invalid")
	}

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

	return historicPoints
}
