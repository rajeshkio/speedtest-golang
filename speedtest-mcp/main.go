package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	ResultOutput string `json:"resultoutput"`
}

type GetAllResultsInput struct {
}
type GetAllResultsOutput struct {
	Results string `json:"results" jsonschema:"the complete result to show to the user"`
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
		ResultOutput: string(jsonSpeedResult),
	}, nil

}

func main() {
	// Create a server with a single tool.
	log.Println("calling from main")
	server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "v1.0.0"}, nil)
	log.Println("before calling the mcp.addtool")
	mcp.AddTool(server, &mcp.Tool{Name: "getAllResults", Description: "returns all the speedtest run"}, getAllResults)
	mcp.AddTool(server, &mcp.Tool{Name: "getSlowSpeedResults", Description: "returns slow speedtest run"}, getSlowSpeedPeriods)
	// Run the server over stdin/stdout, until the client disconnects.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
