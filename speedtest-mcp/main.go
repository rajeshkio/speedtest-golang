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
type GetSlowestPeriodInput struct {
}
type GetSLowestPeriodOutput struct {
	Results string `json:"results"`
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

func getSlowestPeriod(ctx context.Context, req *mcp.CallToolRequest, input GetSlowestPeriodInput) (*mcp.CallToolResult, GetSLowestPeriodOutput, error) {
	results, err := fetchAllResults(url)
	if err != nil {
		return nil, GetSLowestPeriodOutput{}, err
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
		return nil, GetSLowestPeriodOutput{}, err
	}
	return nil, GetSLowestPeriodOutput{
		Results: string(jsonResult),
	}, nil
}

func main() {
	// Create a server with a single tool.
	log.Println("calling from main")
	server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "v1.0.0"}, nil)
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

	// Run the server over stdin/stdout, until the client disconnects.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
