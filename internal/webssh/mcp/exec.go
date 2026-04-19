package mcp

import (
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/webssh/service"
	"bufio"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"log"
	"net/http"
	"strings"
	"time"
)

const commandSentinel = "miku_chan_daisuki"

func CommandTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := runToolCommand(ctx, request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(result), nil
}

func readOutput(reader *bufio.Reader) chan string {
	outputChan := make(chan string)
	go func() {
		defer close(outputChan)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading output:", err)
				return
			}
			outputChan <- line
		}
	}()

	return outputChan
}

func runToolCommand(ctx context.Context, request mcp.CallToolRequest) (string, error) {
	// 这个参数由python程序自动注入，不需要LLM提供
	sessionID, err := request.RequireString("session_id")
	if err != nil {
		return "", err
	}
	command, err := request.RequireString("command")
	if err != nil {
		return "", err
	}
	timeout, err := request.RequireInt("timeout")
	if err != nil {
		return "", err
	}

	session, exists := service.GetSession(sessionID)
	if !exists {
		return "", fmt.Errorf("session not found")
	}

	sideBuffer := session.SetSideBuffer()
	defer session.ClearSideBuffer()

	if err := session.WriteInput([]byte(command + "; echo '" + commandSentinel + "'" + "\n")); err != nil {
		return "", err
	}

	reader := bufio.NewReader(sideBuffer)
	result := ""
	timer := time.After(time.Duration(timeout) * time.Second)
	outputCh := readOutput(reader)
	for {
		select {
		case <-timer:
			return "", fmt.Errorf("Command execution timed out, the latest output is: \n%s", result)
		case <-ctx.Done():
			return "", fmt.Errorf("Command execution cancelled, the latest output is: \n%s", result)
		case line, ok := <-outputCh:
			if !ok {
				return result, nil
			}
			fmt.Println(line)
			if strings.TrimSpace(line) == commandSentinel {
				return result, nil
			}
			result += line
		}
	}
}

func StartMCP() {
	s := server.NewMCPServer(
		"SSH Executor",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	// Add an executor tool
	executorTool := mcp.NewTool("command_executor",
		mcp.WithDescription("Executes a command and returns the output"),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("The command to execute"),
		),
		mcp.WithNumber("timeout",
			mcp.Required(),
			mcp.Description("Execution timeout in seconds"),
		),
	)

	// Add the handler
	s.AddTool(executorTool, CommandTool)

	// Start StreamableHTTP server
	mcpPort := config.Get().MCPPort
	log.Println(fmt.Sprintf("Starting StreamableHTTP server on :%d", mcpPort))
	httpServer := server.NewStreamableHTTPServer(
		s,
		server.WithStateful(true),
		server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			return ctx
		}),
	)
	if err := httpServer.Start(fmt.Sprintf(":%d", mcpPort)); err != nil {
		log.Fatal(err)
	}
}
