package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const url = "http://127.0.0.1:8010"

type GetSlowPeriodsInput struct {
	SpeedThreshold float64 `json:"speedthreshold"`
}

type GetSlowPeriodsOutput struct {
	Results string `json:"results"`
}

type GetAllResultsInput struct {
}
type GetAllResultsOutput struct {
	Results string `json:"results" jsonschema:"the complete result to show to the user"`
}

// no input needed — always returns the single slowest
type GetSlowestPeriodInput struct {
}
type GetSlowestPeriodOutput struct {
	Results string `json:"results"`
}

// needs date input — user specifies which date
type GetResultByDateInput struct {
	Date string `json:"date"`
}
type GetResultByDateOutput struct {
	Results string `json:"results"`
}

type GetResultsByDurationInput struct {
	Duration string `json:"duration"`
}

type GetResultsByDurationOutput struct {
	Results []SpeedTestResult `json:"results"`
}

type SpeedTestResult struct {
	ID            int64     `json:"ID"`
	TimeStamp     time.Time `json:"TimeStamp"`
	DownloadSpeed string    `json:"DownloadSpeed"`
	UploadSpeed   string    `json:"UploadSpeed"`
	Latency       string    `json:"Latency"`
	PublicIP      string    `json:"PublicIp"`
	ISP           string    `json:"ISP"`
	Peers         string    `json:"Peers"`
}

func fetchAllResults(url string) ([]SpeedTestResult, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []SpeedTestResult{}, fmt.Errorf("http call failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []SpeedTestResult{}, fmt.Errorf("failed to read the response body: %w", err)
	}

	var speedResult []SpeedTestResult
	err = json.Unmarshal(body, &speedResult)
	if err != nil {
		return []SpeedTestResult{}, fmt.Errorf("failed to parse the response body: %w", err)
	}
	return speedResult, nil
}
func getAllResults(ctx context.Context, req *mcp.CallToolRequest, input GetAllResultsInput) (*mcp.CallToolResult, GetAllResultsOutput, error) {
	speedResults, err := fetchAllResults(url)
	if err != nil {
		return nil, GetAllResultsOutput{}, err
	}
	speedResultByte, err := json.Marshal(speedResults)
	if err != nil {
		return nil, GetAllResultsOutput{}, err
	}

	return nil, GetAllResultsOutput{Results: string(speedResultByte)}, nil
}

func getSlowSpeedPeriods(ctx context.Context, req *mcp.CallToolRequest, input GetSlowPeriodsInput) (*mcp.CallToolResult, GetSlowPeriodsOutput, error) {
	results, err := fetchAllResults(url)
	if err != nil {
		return nil, GetSlowPeriodsOutput{}, err
	}
	var slowSpeedResults []SpeedTestResult
	for i, result := range results {
		downloadSpeedFloat, err := strconv.ParseFloat(result.DownloadSpeed, 64)
		if err != nil {
			log.Printf("invalid download speed: %v", err)
		}
		if downloadSpeedFloat <= input.SpeedThreshold {
			slowSpeedResults = append(slowSpeedResults, results[i])
		}
	}

	jsonSpeedResult, err := json.Marshal(slowSpeedResults)
	if err != nil {
		return nil, GetSlowPeriodsOutput{}, fmt.Errorf("failed to marshal: %w", err)
	}
	return nil, GetSlowPeriodsOutput{
		Results: string(jsonSpeedResult),
	}, nil
}

func getSlowestPeriod(ctx context.Context, req *mcp.CallToolRequest, input GetSlowestPeriodInput) (*mcp.CallToolResult, GetSlowestPeriodOutput, error) {
	results, err := fetchAllResults(url)
	if err != nil {
		return nil, GetSlowestPeriodOutput{}, err
	}

	var slowestResult SpeedTestResult
	slowestSpeedResult := math.MaxFloat64
	for _, result := range results {
		speed, err := strconv.ParseFloat(result.DownloadSpeed, 64)
		if err != nil {
			continue
		}
		if speed < slowestSpeedResult {
			slowestSpeedResult = speed
			slowestResult = result
		}
	}

	jsonResult, err := json.Marshal([]SpeedTestResult{slowestResult})
	if err != nil {
		return nil, GetSlowestPeriodOutput{}, err
	}
	return nil, GetSlowestPeriodOutput{
		Results: string(jsonResult),
	}, nil
}

func getResultsByDate(ctx context.Context, req *mcp.CallToolRequest, input GetResultByDateInput) (*mcp.CallToolResult, GetResultByDateOutput, error) {
	results, err := fetchAllResults(url)
	if err != nil {
		return nil, GetResultByDateOutput{}, err
	}

	var resultsByDate []SpeedTestResult
	for _, result := range results {
		if result.TimeStamp.Format("2006-01-02") == input.Date {
			resultsByDate = append(resultsByDate, result)
		}
	}

	jsonResultByDate, err := json.Marshal(resultsByDate)
	if err != nil {
		return nil, GetResultByDateOutput{}, fmt.Errorf("failed to marshal: %w", err)
	}

	return nil, GetResultByDateOutput{
		Results: string(jsonResultByDate),
	}, nil

}

func parseDuration(s string) (time.Duration, error) {

	if strings.HasSuffix(s, "d") {
		numStr := strings.TrimSuffix(s, "d")
		num, err := strconv.Atoi(numStr)
		if err != nil {
			log.Printf("failed to convert duration: %v", s)
			return 0, err
		}
		return time.Duration(num) * 24 * time.Hour, nil
	}

	if strings.HasSuffix(s, "w") {
		numStr := strings.TrimSuffix(s, "w")
		num, err := strconv.Atoi(numStr)
		if err != nil {
			log.Printf("failed to convert duration: %v", s)
			return 0, err
		}
		return time.Duration(num) * 7 * 24 * time.Hour, nil
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	return duration, nil

}

func getResultsByDuration(ctx context.Context, req *mcp.CallToolRequest, input GetResultsByDurationInput) (*mcp.CallToolResult, GetResultsByDurationOutput, error) {
	results, err := fetchAllResults(url)
	if err != nil {
		return nil, GetResultsByDurationOutput{}, err
	}

	duration, err := parseDuration(input.Duration)
	if err != nil {
		return nil, GetResultsByDurationOutput{}, err
	}
	cutoff := time.Now().UTC().Add(-duration)
	filteredResults := []SpeedTestResult{}
	for _, result := range results {
		if result.TimeStamp.After(cutoff) || result.TimeStamp.Equal(cutoff) {
			filteredResults = append(filteredResults, result)
		}
	}

	return nil, GetResultsByDurationOutput{
		Results: filteredResults,
	}, nil
}

func main() {
	// Create a server with a single tool.
	log.Println("calling from main")
	server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "1.0.0"}, nil)
	log.Println("before calling the mcp.addtool")
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getAllResults",
		Description: "Returns all speedtest results including timestamp, download speed, upload speed, latency, ISP and peer. Use this when the user asks for averages, trends, comparisons, best/worst periods, or any question that requires analyzing the full dataset.",
	}, getAllResults)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getSlowSpeedResults",
		Description: "Returns only speedtest results where download speed is below a given threshold in Mbps. Use this when the user asks for slow periods, bad speeds, or speeds below a specific value. Requires speedthreshold parameter.",
	}, getSlowSpeedPeriods)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getSlowestPeriod",
		Description: "Returns the single speedtest result with the lowest download speed ever recorded. Use this when the user asks for the worst period, slowest connection, minimum download speed, or when the internet was slowest. Returns one record with timestamp and speed.",
	}, getSlowestPeriod)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getResultsByDate",
		Description: "Returns all speedtest results for a specific date. Use this when the user asks for results on a specific day, today's results, or yesterday's results. Requires date parameter in YYYY-MM-DD format.",
	}, getResultsByDate)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "getResultsByDuration",
		Description: "Returns speedtest results from the last N duration. Use when the user asks about recent results, current performance, or what happened in the last Xa hours/days. Accepts duration in formats: 30m, 2h, 1d, 2d, 1w. Requires duration parameter.",
	}, getResultsByDuration)

	// Create the streamable HTTP handler.
	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return server
	}, nil)

	// Start the HTTP server.
	if err := http.ListenAndServe("0.0.0.0:8020", handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
