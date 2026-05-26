package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	tuiCmd := &cobra.Command{
		Use:   "tui [-- args...]",
		Short: "Launch the Kode terminal UI",
		Long: `Launch the interactive Kode terminal user interface.

On first run, the opencode compiled binary is downloaded from GitHub.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return launchTUI(args)
		},
	}
	rootCmd.AddCommand(tuiCmd)
}

func launchTUI(args []string) error {
	proxy := findOpencodeBinary()
	if proxy == "" {
		var err error
		proxy, err = downloadOpencodeBinary()
		if err != nil {
			return fmt.Errorf("TUI not available: %w\nInstall with: npm install -g opencode-ai", err)
		}
	}

	selfPath, _ := os.Executable()
	tuiCmd := exec.Command(proxy, args...)
	tuiCmd.Stdin = os.Stdin
	tuiCmd.Stdout = os.Stdout
	tuiCmd.Stderr = os.Stderr
	tuiCmd.Env = append(os.Environ(), fmt.Sprintf("KODE_BIN=%s", selfPath))

	if err := tuiCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("TUI exited: %w", err)
	}
	return nil
}

func findOpencodeBinary() string {
	// Priority 1: Our compiled Kode binary in ~/.kode/tui/bin/
	if home, err := os.UserHomeDir(); err == nil {
		bin := filepath.Join(home, ".kode", "tui", "bin", kodeTuiBinName())
		if info, err := os.Stat(bin); err == nil && !info.IsDir() {
			if versionFileMatches(filepath.Join(home, ".kode", "tui", ".kode-tui-version")) {
				return bin
			}
			os.Remove(bin)
		}
		// Priority 2: npm global install (upstream opencode, Kode-branded fallback)
		if home := os.Getenv("APPDATA"); home != "" {
			bin := filepath.Join(home, "npm", "node_modules", "opencode-ai", "bin", "opencode.exe")
			if info, err := os.Stat(bin); err == nil && !info.IsDir() {
				return bin
			}
		}
		// Unix npm
		if home, err := os.UserHomeDir(); err == nil {
			for _, p := range []string{
				filepath.Join(home, ".npm", "packages", "opencode-ai", "bin", "opencode"),
				filepath.Join(home, ".local", "share", "npm", "opencode-ai", "bin", "opencode"),
			} {
				if info, err := os.Stat(p); err == nil && !info.IsDir() {
					return p
				}
			}
		}
	}
	return ""
}

func kodeTuiBinName() string {
	if runtime.GOOS == "windows" {
		return "kode-tui.exe"
	}
	return "kode-tui"
}

func downloadOpencodeBinary() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	tuiDir := filepath.Join(homeDir, ".kode", "tui")
	binDir := filepath.Join(tuiDir, "bin")
	os.MkdirAll(binDir, 0755)

	tag := version
	if tag == "" || tag == "dev" || tag == "none" {
		tag = "latest"
	}

	url := kodeTuiAssetURL(tag)
	binPath := filepath.Join(binDir, kodeTuiBinName())

	fmt.Fprintf(os.Stderr, "Downloading Kode TUI binary from GitHub...\n")
	tmpPath := binPath + ".download"
	if err := downloadBinary(url, tmpPath); err != nil {
		return "", fmt.Errorf("download Kode TUI binary: %w", err)
	}
	os.Chmod(tmpPath, 0755)
	os.Rename(tmpPath, binPath)
	os.WriteFile(filepath.Join(tuiDir, ".kode-tui-version"), []byte(version), 0644)

	return binPath, nil
}

func kodeTuiAssetURL(tag string) string {
	v := "1.15.10"
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x64"
	}
	asset := fmt.Sprintf("kode-tui-%s-%s", runtime.GOOS, arch)
	if runtime.GOOS == "windows" {
		asset += ".exe"
	}
	return fmt.Sprintf("https://github.com/sicario-labs/kode/releases/download/v%s/%s", version, asset)
}

func versionFileMatches(path string) bool {
	if version == "" || version == "dev" || version == "none" {
		return true
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == version
}

func downloadBinary(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
