package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const tuiBundleFile = "tui-bundle.tar.gz"

func findTUIDir() (string, error) {
	if env := os.Getenv("KODE_TUI_DIR"); env != "" {
		if info, err := os.Stat(env); err == nil && info.IsDir() {
			return env, nil
		}
	}

	selfPath, err := os.Executable()
	searchDirs := []string{}

	if err == nil {
		selfDir := filepath.Dir(selfPath)
		// Look for TUI relative to the binary location (installed layout)
		searchDirs = append(searchDirs,
			filepath.Join(selfDir, ".."),
		)
	}

	cwd, _ := os.Getwd()
	if cwd != "" {
		// Dev mode: TUI is at the repo root
		searchDirs = append(searchDirs, cwd)
	}

	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		searchDirs = append(searchDirs, filepath.Join(homeDir, ".kode", "tui"))
	}

	bundleDir := ""
	if homeDir != "" {
		bundleDir = filepath.Join(homeDir, ".kode", "tui")
	}

	for _, dir := range searchDirs {
		abs, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if info, statErr := os.Stat(abs); statErr == nil && info.IsDir() {
			if bundleDir != "" && abs == bundleDir {
				match, _ := versionMatches(abs)
				if !match {
					continue
				}
			}
			return abs, nil
		}
	}

	return "", fmt.Errorf("TUI directory not found")
}

func versionMatches(tuiDir string) (bool, error) {
	if version == "" || version == "dev" || version == "none" {
		return true, nil
	}
	data, err := os.ReadFile(filepath.Join(tuiDir, ".kode-version"))
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(data)) == version, nil
}

func ensureTUI() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	kodeDir := filepath.Join(homeDir, ".kode")
	tuiDir := filepath.Join(kodeDir, "tui")

	if info, err := os.Stat(tuiDir); err == nil && info.IsDir() {
		match, _ := versionMatches(tuiDir)
		if match {
			return tuiDir, nil
		}
		os.RemoveAll(tuiDir)
	}

	tag := version
	if tag == "" || tag == "dev" || tag == "none" {
		tag = "latest"
	}

	var url string
	if tag == "latest" {
		url = fmt.Sprintf("https://github.com/sicario-labs/kode/releases/latest/download/%s", tuiBundleFile)
	} else {
		v := strings.TrimPrefix(tag, "v")
		url = fmt.Sprintf("https://github.com/sicario-labs/kode/releases/download/v%s/%s", v, tuiBundleFile)
	}

	if env := os.Getenv("KODE_TUI_BUNDLE_URL"); env != "" {
		url = env
	}

	fmt.Fprintf(os.Stderr, "Downloading TUI bundle (~52 MB) from GitHub Releases...\n")

	if err := os.MkdirAll(kodeDir, 0755); err != nil {
		return "", fmt.Errorf("create kode dir: %w", err)
	}

	if err := downloadAndExtract(url, kodeDir); err != nil {
		os.RemoveAll(tuiDir)
		return "", fmt.Errorf("download TUI: %w", err)
	}

	fmt.Fprintf(os.Stderr, "TUI bundle extracted to %s\n", tuiDir)
	return tuiDir, nil
}

func downloadAndExtract(url, destDir string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP GET: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar: %w", err)
		}

		parts := strings.SplitN(header.Name, "/", 2)
		if len(parts) < 2 || parts[0] != "tui" {
			continue
		}
		rel := parts[1]
		if rel == "" {
			continue
		}

		target := filepath.Join(destDir, "tui", rel)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("mkdir %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("mkdir %s: %w", target, err)
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("create %s: %w", target, err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("write %s: %w", target, err)
			}
			f.Close()
		}
	}

	return nil
}
