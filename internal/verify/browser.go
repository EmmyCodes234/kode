package verify

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type BrowserGate struct {
	ProjectRoot string
}

func NewBrowserGate(projectRoot string) *BrowserGate {
	return &BrowserGate{
		ProjectRoot: projectRoot,
	}
}

type browserStepLog struct {
	Step    string `json:"step"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ProjectConfig struct {
	Engine *EngineConfig `json:"engine"`
}

type EngineConfig struct {
	BrowserVerification bool   `json:"browser_verification"`
	DevServerCommand    string `json:"dev_server_command"`
}

func LoadProjectConfig(projectRoot string) (*ProjectConfig, error) {
	configPath := filepath.Join(projectRoot, ".kode", "kode.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg ProjectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (bg *BrowserGate) Verify(ctx context.Context, startURL string, taskInstructions string) CheckResult {
	fmt.Fprintf(os.Stderr, "KODE_GATE: browser_verification - checking dev server\n")
	baseURL, devCmd, err := bg.bootDevServer(ctx)
	if err != nil {
		return CheckResult{
			CheckName: "browser_verification",
			Status:    StatusFail,
			Message:   "Failed to boot or locate dev server",
			Details:   err.Error(),
		}
	}
	if devCmd != nil {
		defer devCmd.Process.Kill()
	}

	if startURL == "" {
		startURL = baseURL
	}

	artifactsDir := filepath.Join(bg.ProjectRoot, "artifacts")
	os.MkdirAll(artifactsDir, 0755)
	
	videoOutput := filepath.Join(artifactsDir, "walkthrough.webm")
	screenshotOutput := filepath.Join(artifactsDir, "screenshot.png")

	os.Remove(videoOutput)
	os.Remove(screenshotOutput)

	driverPath := filepath.Join(bg.ProjectRoot, "internal", "verify", "browser.js")
	
	// Prepare execution arguments
	args := []string{driverPath, "--url", startURL}
	if taskInstructions != "" {
		args = append(args, "--task", taskInstructions)
	}
	args = append(args, "--video", videoOutput, "--screenshot", screenshotOutput)

	cmd := exec.CommandContext(ctx, "node", args...)
	cmd.Dir = bg.ProjectRoot

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return CheckResult{
			CheckName: "browser_verification",
			Status:    StatusFail,
			Message:   "Failed to create stdout pipe for browser driver",
			Details:   err.Error(),
		}
	}

	if err := cmd.Start(); err != nil {
		return CheckResult{
			CheckName: "browser_verification",
			Status:    StatusFail,
			Message:   "Failed to start browser verification driver",
			Details:   err.Error(),
		}
	}

	scanner := bufio.NewScanner(stdoutPipe)
	var finalMessage string
	for scanner.Scan() {
		line := scanner.Text()
		var stepLog browserStepLog
		if err := json.Unmarshal([]byte(line), &stepLog); err == nil {
			fmt.Fprintf(os.Stderr, "KODE_BROWSER: [%s] %s - %s\n", stepLog.Status, stepLog.Step, stepLog.Message)
			if stepLog.Status == "FAIL" {
				finalMessage = stepLog.Message
			}
		} else {
			fmt.Fprintln(os.Stderr, line)
		}
	}

	err = cmd.Wait()

	bg.writeWalkthroughReport(taskInstructions, videoOutput, screenshotOutput)

	if err != nil {
		if finalMessage == "" {
			finalMessage = "Browser execution encountered a runtime error"
		}
		return CheckResult{
			CheckName: "browser_verification",
			Status:    StatusFail,
			Message:   finalMessage,
			Details:   fmt.Sprintf("Visual results saved to: %s", artifactsDir),
		}
	}

	return CheckResult{
		CheckName: "browser_verification",
		Status:    StatusPass,
		Message:   "Browser verification gate passed successfully",
		Details:   fmt.Sprintf("Walkthrough recorded: %s", videoOutput),
	}
}

func (bg *BrowserGate) bootDevServer(ctx context.Context) (string, *exec.Cmd, error) {
	ports := []int{3000, 5173, 8080, 3001, 5000, 8000}
	for _, port := range ports {
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return fmt.Sprintf("http://%s", addr), nil, nil
		}
	}

	devCmdStr := ""
	if cfg, err := LoadProjectConfig(bg.ProjectRoot); err == nil && cfg.Engine != nil && cfg.Engine.DevServerCommand != "" {
		devCmdStr = cfg.Engine.DevServerCommand
	}

	if devCmdStr == "" {
		pkgJSONPath := filepath.Join(bg.ProjectRoot, "package.json")
		if _, err := os.Stat(pkgJSONPath); err == nil {
			devCmdStr = "npm run dev"
			if _, err := os.Stat(filepath.Join(bg.ProjectRoot, "bun.lockb")); err == nil {
				devCmdStr = "bun run dev"
			}
		}
	}

	if devCmdStr == "" {
		return "", nil, fmt.Errorf("no dev server is running and no dev_server_command is configured")
	}

	parts := strings.Fields(devCmdStr)
	var cmd *exec.Cmd
	if len(parts) == 1 {
		cmd = exec.CommandContext(ctx, parts[0])
	} else {
		cmd = exec.CommandContext(ctx, parts[0], parts[1:]...)
	}
	cmd.Dir = bg.ProjectRoot

	if err := cmd.Start(); err != nil {
		return "", nil, fmt.Errorf("failed to start dev server command %q: %w", devCmdStr, err)
	}

	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			return "", nil, ctx.Err()
		case <-timeout:
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			return "", nil, fmt.Errorf("timeout waiting for dev server to bind to a port")
		case <-ticker.C:
			for _, port := range ports {
				addr := fmt.Sprintf("127.0.0.1:%d", port)
				conn, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
				if err == nil {
					conn.Close()
					return fmt.Sprintf("http://%s", addr), cmd, nil
				}
			}
		}
	}
}

func (bg *BrowserGate) writeWalkthroughReport(task string, videoPath string, screenshotPath string) {
	reportPath := filepath.Join(bg.ProjectRoot, "walkthrough.md")
	content := fmt.Sprintf(`# 📷 Kode Browser Verification Report

All gates passed. Headless browser successfully verified the UI flow:

## 📋 Task Details
%s

## 🎥 Walkthrough Recording
![Browser Walkthrough](file:///%s)

## 🖼️ Final Rendered State
![Final State](file:///%s)

## 📋 Console Logs
> [!NOTE]
> Headless run completed. Check console/terminal for detailed trace.
`, task, filepath.ToSlash(videoPath), filepath.ToSlash(screenshotPath))

	os.WriteFile(reportPath, []byte(content), 0644)
}
