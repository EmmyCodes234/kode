package install

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	githubOwner = "sicario-labs"
	githubRepo  = "sicario-cli"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func latestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", githubOwner, githubRepo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func assetSuffix() string {
	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return "linux-amd64.tar.gz"
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			return "darwin-amd64.tar.gz"
		case "arm64":
			return "darwin-arm64.tar.gz"
		}
	case "windows":
		if runtime.GOARCH == "amd64" {
			return "windows-amd64.zip"
		}
	}
	return ""
}

func EnsureInstalled(targetDir string) (string, error) {
	if p := os.Getenv("SICARIO_PATH"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	if _, err := os.Stat(filepath.Join(targetDir, "sicario")); err == nil {
		return filepath.Join(targetDir, "sicario"), nil
	}
	if _, err := os.Stat(filepath.Join(targetDir, "sicario.exe")); err == nil {
		return filepath.Join(targetDir, "sicario.exe"), nil
	}

	return downloadAndInstall(targetDir)
}

func downloadAndInstall(targetDir string) (string, error) {
	release, err := latestRelease()
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest Sicario release: %w", err)
	}

	suffix := assetSuffix()
	if suffix == "" {
		return "", fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	var assetURL string
	for _, a := range release.Assets {
		if strings.HasSuffix(a.Name, suffix) {
			assetURL = a.BrowserDownloadURL
			break
		}
	}
	if assetURL == "" {
		return "", fmt.Errorf("no asset found for %s (release %s)", suffix, release.TagName)
	}

	resp, err := http.Get(assetURL)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response failed: %w", err)
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("create target dir: %w", err)
	}

	var binaryName string
	if strings.HasSuffix(suffix, ".zip") {
		binaryName, err = extractZip(data, targetDir)
	} else {
		binaryName, err = extractTarGz(data, targetDir)
	}
	if err != nil {
		return "", fmt.Errorf("extract failed: %w", err)
	}

	binPath := filepath.Join(targetDir, binaryName)
	if runtime.GOOS != "windows" {
		os.Chmod(binPath, 0755)
	}
	return binPath, nil
}

func extractTarGz(data []byte, targetDir string) (string, error) {
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var extracted string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		name := filepath.Base(hdr.Name)
		if !strings.HasPrefix(name, "sicario") {
			continue
		}
		outPath := filepath.Join(targetDir, "sicario")
		f, err := os.Create(outPath)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return "", err
		}
		f.Close()
		extracted = "sicario"
	}
	if extracted == "" {
		return "", fmt.Errorf("no sicario binary found in archive")
	}
	return extracted, nil
}

func extractZip(data []byte, targetDir string) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	var extracted string
	for _, f := range zr.File {
		name := filepath.Base(f.Name)
		if !strings.HasPrefix(name, "sicario") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		outPath := filepath.Join(targetDir, "sicario.exe")
		outFile, err := os.Create(outPath)
		if err != nil {
			rc.Close()
			return "", err
		}
		if _, err := io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			return "", err
		}
		rc.Close()
		outFile.Close()
		extracted = "sicario.exe"
	}
	if extracted == "" {
		return "", fmt.Errorf("no sicario binary found in archive")
	}
	return extracted, nil
}
