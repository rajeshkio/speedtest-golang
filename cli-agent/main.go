package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const url = "http://192.168.90.100:11434/api/chat"

type Message struct {
	Role     string     `json:"role"`
	Content  string     `json:"content"`
	ToolCall []ToolCall `json:"tool_calls"`
}
type ChatRequest struct {
	Model    string                 `json:"model"`
	Messages []Message              `json:"messages"`
	Stream   bool                   `json:"stream"`
	Tools    []Tool                 `json:"tools"`
	Think    bool                   `json:"think"`
	Options  map[string]interface{} `json:"options"`
}

type ChatResponse struct {
	Messages Message `json:"message"`
}

type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type ToolCallFunction struct {
	Index     int64                  `json:"index"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}
type ToolCall struct {
	ID       string           `json:"id"`
	Function ToolCallFunction `json:"function"`
}
type MCPToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}
type MCPToolCall struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  MCPToolParams `json:"params"`
}

type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type MCPResult struct {
	Content []MCPContent `json:"content"`
}

type MCPResponse struct {
	Result MCPResult `json:"result"`
}

type SpeedTestResult struct {
	ID            int64  `json:"id"`
	TimeStamp     string `json:"timestamp"`
	DownloadSpeed string `json:"downloadspeed"`
}

type ToolOutput struct {
	ResultOutput string `json:"resultoutput"`
}

func main() {

	//Launch MCP server as subprocess
	cmd := exec.Command("/Users/rajeshkumar/mcp-workspace/speedtest-golang/speedtest-mcp/speedtest-mcp")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("failed to read stdin: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("failed to read stdout: %v", err)
	}
	err = cmd.Start()

	_, err = fmt.Fprintf(stdin, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}`+"\n")
	if err != nil {
		fmt.Printf("failed to write to stdin in mcp initialise: %v", err)
	}
	mcpReader := bufio.NewReader(stdout)
	mcpReader.ReadString('\n')
	_, err = fmt.Fprintf(stdin, `{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}`+"\n")
	if err != nil {
		fmt.Printf("failed to write to stdin in mcp notification: %v", err)
	}

	for {
		fmt.Printf("Ask: ")

		userInputReader := bufio.NewReader(os.Stdin)
		input, _ := userInputReader.ReadString('\n')
		if strings.TrimSpace(input) == "exit" {
			break
		}
		// Talk to Ollama, get a reply
		chatRequest := ChatRequest{
			Model: "qwen2.5:14b",
			Messages: []Message{
				Message{Role: "system", Content: "You are a data extraction assistant. Return ONLY the exact fields requested. Format each record as: Timestamp | Download Speed Mbps. One record per line. No ID, no JSON, no arrays, no analysis, no headers, no extra commentary."},
				Message{Role: "user", Content: input},
			},
			Stream: false,
			Think:  false,
			Options: map[string]interface{}{
				"num_ctx": 16384,
			},
			// Send tool definitions, Ollama decides to call a tool
			Tools: []Tool{Tool{
				Type: "function",
				Function: Function{
					Name:        "getSlowSpeedResults",
					Description: "Returns periods where download speed was below a threshold",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"speedthreshold": map[string]interface{}{
								"type":        "number",
								"description": "Download speed threshold in Mbps",
							},
						},
						"required": []string{"speedthreshold"},
					},
				},
			},
			},
		}

		chatRequestByte, err := json.Marshal(chatRequest)
		if err != nil {
			log.Printf("failed to marshal the chatRequest: %v", err)
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(chatRequestByte))
		if err != nil {
			log.Printf("failed to create request: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("failed to do the request")
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("failed to read the response: %v", err)
		}
		resp.Body.Close()
		var chatResponse ChatResponse
		err = json.Unmarshal(body, &chatResponse)
		if err != nil {
			fmt.Printf("failed to unmarshall the response body: %v", err)
		}

		// Detect tool call vs plain text answer
		if len(chatResponse.Messages.ToolCall) > 0 {
			for _, tool := range chatResponse.Messages.ToolCall {
				toolByte, err := json.Marshal(tool.Function.Arguments)
				if err != nil {
					fmt.Printf("failed to marshal function argument")
				}
				MCPRequest := MCPToolCall{
					JSONRPC: "2.0",
					ID:      3,
					Method:  "tools/call",
					Params: MCPToolParams{
						Name:      tool.Function.Name,
						Arguments: json.RawMessage(toolByte),
					},
				}

				MCPRequestByte, _ := json.Marshal(MCPRequest)
				fmt.Fprintf(stdin, "%s\n", string(MCPRequestByte))
				line, _ := mcpReader.ReadString('\n')

				var mcpResponse MCPResponse
				err = json.Unmarshal([]byte(line), &mcpResponse)
				if err != nil {
					fmt.Printf("failed to unmarshal mcpresponse: %v", err)
				}

				text := mcpResponse.Result.Content[0].Text
				var toolOutput ToolOutput
				err = json.Unmarshal([]byte(text), &toolOutput)
				if err != nil {
					fmt.Printf("failed to unmarshal mcpresponse content: %v", err)
				}

				var speedTestResult []SpeedTestResult
				json.Unmarshal([]byte(toolOutput.ResultOutput), &speedTestResult)

				for _, result := range speedTestResult {
					fmt.Printf("%s - %s Mbps \n", result.TimeStamp, result.DownloadSpeed)
				}
			}
		} else {
			fmt.Println(chatResponse.Messages.Content)
		}
	}
}
