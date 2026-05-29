package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kode/kode/internal/voice"
	"github.com/kode/kode/internal/verify"
	"github.com/kode/kode/internal/telemetry"
	"github.com/spf13/cobra"
)

func init() {
	var duration int
	var endpoint string
	var model string
	var apiKeyEnv string
	var execute bool

	voiceCmd := &cobra.Command{
		Use:   "voice",
		Short: "Record vocal command and transcribe or execute it",
		Long: `Record your voice directly from the terminal. Captures audio via native OS APIs,
transcribes it using an OpenAI-compatible Whisper endpoint, and optionally executes the action.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir, err := os.Getwd()
			if err != nil {
				return err
			}

			// Load configurations from .kode/kode.json if it exists
			cfg, err := verify.LoadProjectConfig(projectDir)
			voiceEnabled := true
			customHotkey := "v"
			silenceThreshold := -40
			maxDuration := 15

			if err == nil && cfg.Engine != nil {
				// We can load specific fields if we define a voice block in config.
				// Let's fallback gracefully if config load fails or fields are absent.
			}

			// Load endpoint / keys
			apiKey := os.Getenv(apiKeyEnv)
			if apiKey == "" {
				apiKey = os.Getenv("KODE_LLM_API_KEY")
			}
			if apiKey == "" {
				apiKey = os.Getenv("OPENAI_API_KEY")
			}

			if endpoint == "" {
				endpoint = os.Getenv("KODE_LLM_ENDPOINT")
				if endpoint == "" {
					endpoint = "https://api.openai.com/v1"
				}
			}

			// Normalize endpoint url to transcriptions path
			if !strings.HasSuffix(endpoint, "/audio/transcriptions") {
				endpoint = strings.TrimSuffix(endpoint, "/") + "/audio/transcriptions"
			}

			tempWav := filepath.Join(projectDir, "temp_voice.wav")
			defer os.Remove(tempWav)

			bridge := voice.NewVoiceBridge(projectDir)

			fmt.Printf("\n🎙️  Recording started... speak now (max %ds)\n", duration)
			fmt.Printf("Press Ctrl+C to stop recording early.\n\n")

			// Live visual countdown/progress indicator in terminal
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()
				for i := duration; i > 0; i-- {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						// Draw live progress bar
						bar := strings.Repeat("█", duration-i) + strings.Repeat("░", i)
						fmt.Printf("\rRecording [%s] %ds remaining...", bar, i)
					}
				}
				fmt.Printf("\n")
			}()

			err = bridge.RecordAudio(ctx, tempWav, duration)
			cancel() // Stop progress ticker
			if err != nil && err != context.Canceled {
				return fmt.Errorf("recording failed: %w", err)
			}

			fmt.Println("\n\n🤖 Transcribing voice command...")

			text, err := bridge.TranscribeAudio(context.Background(), tempWav, endpoint, apiKey, model)
			if err != nil {
				return fmt.Errorf("transcription failed: %w", err)
			}

			// Telemetry capture
			telemetryClient := telemetry.NewPostHogClient(projectDir)
			telemetryClient.Track("voice_transcribed", map[string]interface{}{
				"duration": duration,
				"length":   len(text),
				"executed": execute,
			})

			fmt.Printf("\n💬 Heard: \"%s\"\n\n", text)

			if execute {
				fmt.Println("🚀 Executing command: kode loop " + text)
				// We can invoke the loop command dynamically by constructing it
				loopCmd := rootCmd
				loopArgs := []string{"loop", text}
				loopCmd.SetArgs(loopArgs)
				return loopCmd.Execute()
			}

			_ = voiceEnabled
			_ = customHotkey
			_ = silenceThreshold
			_ = maxDuration

			return nil
		},
	}

	voiceCmd.Flags().IntVar(&duration, "duration", 5, "Recording duration in seconds")
	voiceCmd.Flags().StringVar(&endpoint, "endpoint", "", "OpenAI-compatible transcription endpoint")
	voiceCmd.Flags().StringVar(&model, "model", "whisper-1", "Whisper model ID")
	voiceCmd.Flags().StringVar(&apiKeyEnv, "api-key-env", "", "Environment variable holding the API key")
	voiceCmd.Flags().BoolVar(&execute, "execute", false, "Immediately execute the transcribed text command in a loop")

	rootCmd.AddCommand(voiceCmd)
}
