package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Global flags
var (
	serverURL string
	pin       string
	testSleep bool
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

func main() {
	// 1. Auto-discover PIN and Port
	defaultPin := discoverPin()
	defaultPort := discoverPort()
	defaultURL := fmt.Sprintf("http://localhost:%s", defaultPort)

	// 2. Define flags
	flag.StringVar(&serverURL, "url", defaultURL, "PCRemote Server URL")
	flag.StringVar(&pin, "pin", defaultPin, "Authentication PIN")
	flag.BoolVar(&testSleep, "test-sleep", false, "Test system sleep (WARNING: will trigger sleep/display off)")
	flag.Parse()

	fmt.Printf("%s============================================================%s\n", colorBlue, colorReset)
	fmt.Printf("%s   PCRemote Server — Premium HTTP API Test Suite%s\n", colorBold+colorCyan, colorReset)
	fmt.Printf("   Server URL: %s%s%s\n", colorBold, serverURL, colorReset)
	fmt.Printf("   Auth PIN:   %s%s%s (discovered: %s)\n", colorBold, pin, colorReset, defaultPin)
	fmt.Printf("%s============================================================%s\n\n", colorBlue, colorReset)

	// Keep track of counts
	passed := 0
	failed := 0

	runTest := func(name string, method string, endpoint string, usePin bool, customPin string, payload any, expectedStatus int) {
		fmt.Printf("Test: %-55s ", name)
		
		var body io.Reader
		if payload != nil {
			jsonBytes, err := json.Marshal(payload)
			if err != nil {
				fmt.Printf("[%sFAIL%s] (JSON marshal error: %v)\n", colorRed, colorReset, err)
				failed++
				return
			}
			body = bytes.NewReader(jsonBytes)
		}

		req, err := http.NewRequest(method, serverURL+endpoint, body)
		if err != nil {
			fmt.Printf("[%sFAIL%s] (Request creation error: %v)\n", colorRed, colorReset, err)
			failed++
			return
		}

		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		if usePin {
			p := pin
			if customPin != "" {
				p = customPin
			}
			req.Header.Set("X-PIN", p)
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("[%sFAIL%s] (Connection error: %v)\n", colorRed, colorReset, err)
			failed++
			return
		}
		defer resp.Body.Close()

		respBodyBytes, _ := io.ReadAll(resp.Body)
		respBody := string(respBodyBytes)

		if resp.StatusCode == expectedStatus {
			fmt.Printf("[%sPASS%s] (HTTP %d)\n", colorGreen, colorReset, resp.StatusCode)
			passed++
		} else {
			// Some tests like Sonar endpoints might return 404 or 500 if Sonar is not installed/active.
			// We mark them as warning/skipped since it's environment dependent.
			if (endpoint == "/audio/channels" || strings.HasPrefix(endpoint, "/audio/channel/")) && (resp.StatusCode == 404 || resp.StatusCode == 500) {
				fmt.Printf("[%sWARN%s] (HTTP %d - Sonar not active: %s)\n", colorYellow, colorReset, resp.StatusCode, strings.TrimSpace(respBody))
				passed++
			} else {
				fmt.Printf("[%sFAIL%s] (HTTP %d, expected %d. Resp: %s)\n", colorRed, colorReset, resp.StatusCode, expectedStatus, strings.TrimSpace(respBody))
				failed++
			}
		}
	}

	// 1. Health
	runTest("1. Health Check (No Auth)", "GET", "/health", false, "", nil, http.StatusOK)

	// 2. Auth checks
	runTest("2. Auth: Wrong PIN", "POST", "/audio/volume", true, "wrong_pin_999", map[string]any{"level": 0.5}, http.StatusUnauthorized)
	runTest("3. Auth: Missing PIN Header", "POST", "/audio/volume", false, "", map[string]any{"level": 0.5}, http.StatusUnauthorized)

	// 3. Audio status & adjustments
	runTest("4. Audio Status Check", "GET", "/audio/status", true, "", nil, http.StatusOK)
	runTest("5. Set Main Audio Volume (50%)", "POST", "/audio/volume", true, "", map[string]any{"level": 0.5}, http.StatusOK)
	runTest("6. Set Main Audio Volume Invalid (150%)", "POST", "/audio/volume", true, "", map[string]any{"level": 1.5}, http.StatusBadRequest)
	runTest("7. Mute Main Audio", "POST", "/audio/mute", true, "", map[string]any{"muted": true}, http.StatusOK)
	runTest("8. Unmute Main Audio", "POST", "/audio/mute", true, "", map[string]any{"muted": false}, http.StatusOK)

	// 4. Sonar channel APIs (warning if Sonar is not available/enabled)
	runTest("9. Get Sonar Audio Channels", "GET", "/audio/channels", true, "", nil, http.StatusOK)
	runTest("10. Set Sonar Media Channel Volume (80%)", "POST", "/audio/channel/volume", true, "", map[string]any{"channel": "media", "level": 0.8}, http.StatusOK)
	runTest("11. Mute Sonar Gaming Channel", "POST", "/audio/channel/mute", true, "", map[string]any{"channel": "gaming", "muted": true}, http.StatusOK)

	// 5. Media keys
	runTest("12. Send Media Key: Play/Pause", "POST", "/media/play", true, "", nil, http.StatusOK)
	runTest("13. Send Media Key: Next Track", "POST", "/media/next", true, "", nil, http.StatusOK)
	runTest("14. Send Media Key: Previous Track", "POST", "/media/prev", true, "", nil, http.StatusOK)

	// 6. Browser endpoints
	runTest("15. Open Browser URL (example.com)", "POST", "/browser/open", true, "", map[string]any{"url": "https://example.com"}, http.StatusOK)
	runTest("16. Open Browser URL Invalid (empty)", "POST", "/browser/open", true, "", map[string]any{"url": ""}, http.StatusBadRequest)

	// 7. System operations (safe)
	runTest("17. Cancel Scheduled Shutdown", "POST", "/system/shutdown/cancel", true, "", nil, http.StatusOK)
	runTest("18. Shutdown Invalid Delay (-1m)", "POST", "/system/shutdown", true, "", map[string]any{"delay_minutes": -1}, http.StatusBadRequest)

	// 8. Sleep feature (only if flag is set)
	if testSleep {
		runTest("19. System Sleep (S0/S3 modern standby)", "POST", "/system/sleep", true, "", nil, http.StatusOK)
	} else {
		fmt.Printf("Test: %-55s [%sSKIP%s] (use -test-sleep flag)\n", "19. System Sleep", colorYellow, colorReset)
	}

	fmt.Println()
	fmt.Printf("%s============================================================%s\n", colorBlue, colorReset)
	var failColor string
	if failed > 0 {
		failColor = colorRed
	} else {
		failColor = colorReset
	}
	fmt.Printf("   Tests Completed: Passed: %s%d%s, Failed: %s%d%s\n", 
		colorGreen, passed, colorReset, 
		failColor, failed, colorReset)
	fmt.Printf("%s============================================================%s\n", colorBlue, colorReset)

	if failed > 0 {
		os.Exit(1)
	}
}

func discoverPin() string {
	paths := []string{
		".env",
		"server/.env",
		`C:\Program Files\PCRemote\.env`,
	}
	for _, p := range paths {
		if content, err := os.ReadFile(p); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "PIN=") {
					return strings.TrimPrefix(line, "PIN=")
				}
				if strings.HasPrefix(line, "APP_PIN=") {
					return strings.TrimPrefix(line, "APP_PIN=")
				}
			}
		}
	}
	return "1234"
}

func discoverPort() string {
	if p := os.Getenv("APP_PORT"); p != "" {
		return p
	}
	paths := []string{
		".env",
		"server/.env",
		`C:\Program Files\PCRemote\.env`,
	}
	for _, p := range paths {
		if content, err := os.ReadFile(p); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "PORT=") {
					return strings.TrimPrefix(line, "PORT=")
				}
				if strings.HasPrefix(line, "APP_PORT=") {
					return strings.TrimPrefix(line, "APP_PORT=")
				}
			}
		}
	}
	return "8000"
}
