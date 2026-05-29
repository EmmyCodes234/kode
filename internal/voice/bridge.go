package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

// Winmm DLL functions for Windows recording
var (
	winmm          = syscall.NewLazyDLL("winmm.dll")
	mciSendStringA = winmm.NewProc("mciSendStringA")
)

type VoiceBridge struct {
	ProjectRoot string
}

func NewVoiceBridge(projectRoot string) *VoiceBridge {
	return &VoiceBridge{
		ProjectRoot: projectRoot,
	}
}

type TranscriptionResponse struct {
	Text string `json:"text"`
}

// RecordAudio captures audio from the default mic for duration seconds and writes it to outputPath
func (vb *VoiceBridge) RecordAudio(ctx context.Context, outputPath string, duration int) error {
	os.MkdirAll(filepath.Dir(outputPath), 0755)

	switch runtime.GOOS {
	case "windows":
		return vb.recordWindows(ctx, outputPath, duration)
	case "linux":
		return vb.recordLinux(ctx, outputPath, duration)
	case "darwin":
		return vb.recordMac(ctx, outputPath, duration)
	default:
		return fmt.Errorf("unsupported operating system for native voice recording: %s", runtime.GOOS)
	}
}

func (vb *VoiceBridge) recordWindows(ctx context.Context, outputPath string, duration int) error {
	// Open waveaudio device
	mciSend("open new type waveaudio alias recsound")
	mciSend("record recsound")

	// Wait for duration or context cancellation
	select {
	case <-ctx.Done():
		mciSend("stop recsound")
		mciSend("close recsound")
		return ctx.Err()
	case <-time.After(time.Duration(duration) * time.Second):
	}

	mciSend("stop recsound")
	saveCmd := fmt.Sprintf("save recsound %s", outputPath)
	mciSend(saveCmd)
	mciSend("close recsound")

	return nil
}

func mciSend(cmd string) {
	cStr := append([]byte(cmd), 0)
	mciSendStringA.Call(uintptr(unsafe.Pointer(&cStr[0])), 0, 0, 0)
}

func (vb *VoiceBridge) recordLinux(ctx context.Context, outputPath string, duration int) error {
	// arecord is standard on ALSA systems (default in almost all Linux distros)
	cmd := exec.CommandContext(ctx, "arecord", "-f", "S16_LE", "-c", "1", "-r", "16000", "-d", fmt.Sprintf("%d", duration), outputPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run arecord: %w (stderr: %s)", err, stderr.String())
	}
	return nil
}

func (vb *VoiceBridge) recordMac(ctx context.Context, outputPath string, duration int) error {
	// Try using native screencapture (audio-only) or sox/rec fallback
	if _, err := exec.LookPath("rec"); err == nil {
		cmd := exec.CommandContext(ctx, "rec", "-r", "16000", "-c", "1", outputPath, "trim", "0", fmt.Sprintf("%d", duration))
		return cmd.Run()
	}
	// Dynamic AppleScript command to record audio using QuickTime
	appleScript := fmt.Sprintf(`
		tell application "QuickTime Player"
			set newAudioRecording to new audio recording
			start newAudioRecording
			delay %d
			stop newAudioRecording
			export document 1 to POSIX file "%s" using settings preset "Audio Only"
			close document 1 saving no
		end tell
	`, duration, outputPath)

	cmd := exec.CommandContext(ctx, "osascript", "-e", appleScript)
	return cmd.Run()
}

// TranscribeAudio sends the WAV file to the transcription endpoint
func (vb *VoiceBridge) TranscribeAudio(ctx context.Context, audioPath string, endpoint string, apiKey string, model string) (string, error) {
	file, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create wave file form part with correct MIME type
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filepath.Base(audioPath)))
	h.Set("Content-Type", "audio/wav")
	part, err := writer.CreatePart(h)
	if err != nil {
		return "", fmt.Errorf("failed to create multipart form field: %w", err)
	}

	if _, err = io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy audio file contents: %w", err)
	}

	// Model parameter
	if model == "" {
		model = "whisper-1"
	}
	err = writer.WriteField("model", model)
	if err != nil {
		return "", fmt.Errorf("failed to write model field: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, body)
	if err != nil {
		return "", fmt.Errorf("failed to create POST request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("transcription failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var transResp TranscriptionResponse
	if err := json.Unmarshal(respBody, &transResp); err != nil {
		return "", fmt.Errorf("failed to parse transcription JSON response: %w", err)
	}

	return transResp.Text, nil
}
