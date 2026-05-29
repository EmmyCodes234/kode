package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
)

type TunnelInfo struct {
	URL    string `json:"url"`
	Port   int    `json:"port"`
	Active bool   `json:"active"`
}

func init() {
	tunnelCmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Start a secure preview tunnel for the local web server",
		Long: `Dynamically boots an encrypted preview tunnel using native OpenSSH,
generating a secure URL, writing sync state to .kode/tunnel.json,
and rendering a clean ASCII QR code in the terminal for instant mobile responsive testing.`,
		Run: func(cmd *cobra.Command, args []string) {
			port, _ := cmd.Flags().GetInt("port")

			// Ensure .kode directory exists
			err := os.MkdirAll(".kode", 0755)
			if err != nil {
				fmt.Printf("❌ Failed to create .kode directory: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("🚀 Starting native OpenSSH tunnel forwarding 127.0.0.1:%d to localhost.run...\n", port)

			// Setup signals to cleanly shut down
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			// Start the ssh command:
			// ssh -o StrictHostKeyChecking=no -R 80:127.0.0.1:port nokey@localhost.run
			sshArgs := []string{
				"-o", "StrictHostKeyChecking=no",
				"-R", fmt.Sprintf("80:127.0.0.1:%d", port),
				"nokey@localhost.run",
			}
			
			sshCmd := exec.Command("ssh", sshArgs...)
			
			stdout, err := sshCmd.StdoutPipe()
			if err != nil {
				fmt.Printf("❌ Failed to get stdout pipe: %v\n", err)
				os.Exit(1)
			}
			
			stderr, err := sshCmd.StderrPipe()
			if err != nil {
				fmt.Printf("❌ Failed to get stderr pipe: %v\n", err)
				os.Exit(1)
			}

			err = sshCmd.Start()
			if err != nil {
				fmt.Printf("❌ Failed to start ssh process (ensure OpenSSH is installed and on your PATH): %v\n", err)
				os.Exit(1)
			}

			// Clean up function
			cleanup := func() {
				if sshCmd.Process != nil {
					_ = sshCmd.Process.Kill()
				}
				_ = os.Remove(".kode/tunnel.json")
			}

			// Watch for interruptions in a goroutine
			go func() {
				<-sigChan
				fmt.Println("\n🛑 Terminating secure tunnel and cleaning up sync state...")
				cleanup()
				os.Exit(0)
			}()

			defer cleanup()

			// Read stdout to find the HTTPS URL
			scanner := bufio.NewScanner(stdout)
			re := regexp.MustCompile(`https://[a-zA-Z0-9.-]+\.lhr\.life`)
			
			urlChan := make(chan string, 1)
			errChan := make(chan error, 1)

			go func() {
				for scanner.Scan() {
					line := scanner.Text()
					match := re.FindString(line)
					if match != "" {
						urlChan <- match
						return
					}
				}
				if err := scanner.Err(); err != nil {
					errChan <- err
				} else {
					errChan <- fmt.Errorf("SSH tunnel closed before generating a URL")
				}
			}()

			// Stderr scanner to print errors in case of connection failure
			go func() {
				errScanner := bufio.NewScanner(stderr)
				for errScanner.Scan() {
					// Discard or log stderr if needed
				}
			}()

			var tunnelURL string
			select {
			case tunnelURL = <-urlChan:
				// Found URL!
			case err := <-errChan:
				fmt.Printf("❌ SSH tunnel error: %v\n", err)
				return
			case <-time.After(15 * time.Second):
				fmt.Println("❌ SSH tunnel connection timed out (15s)")
				return
			}

			// Write sync state file .kode/tunnel.json
			info := TunnelInfo{
				URL:    tunnelURL,
				Port:   port,
				Active: true,
			}
			infoBytes, err := json.MarshalIndent(info, "", "  ")
			if err == nil {
				_ = os.WriteFile(".kode/tunnel.json", infoBytes, 0644)
			}

			fmt.Printf("📡 Connected! Tunneling localhost:%d -> %s\n\n", port, tunnelURL)
			fmt.Println("📷 Scan the QR code below to preview on your mobile device:")
			fmt.Println()
			renderASCIIQRCode(tunnelURL)
			fmt.Println()
			fmt.Printf("🔗 Secure Web URL: \033[1;36m%s\033[0m\n", tunnelURL)
			fmt.Println("💡 Press Ctrl+C to terminate the tunnel.")

			if os.Getenv("KODE_TUNNEL_TEST") == "1" {
				return
			}

			// Wait for the ssh command to exit (will block until Ctrl+C)
			_ = sshCmd.Wait()
		},
	}

	tunnelCmd.Flags().IntP("port", "p", 3000, "Local port to tunnel")
	rootCmd.AddCommand(tunnelCmd)
}

func renderASCIIQRCode(url string) {
	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		fmt.Printf("⚠️ Unable to generate dynamic QR code: %v\n", err)
		return
	}

	matrix := qr.Bitmap()
	
	// Print with a light quiet-zone border for optimal scanning contrast
	fmt.Println("    " + strings.Repeat("██", len(matrix)+4))
	
	for _, row := range matrix {
		fmt.Print("  ████  ") // Left margin quiet zone
		for _, cell := range row {
			if cell {
				fmt.Print("  ") // Dark modules (black in QR)
			} else {
				fmt.Print("██") // Light background (white in QR, using foreground color)
			}
		}
		fmt.Println("  ████")
	}
	
	fmt.Println("    " + strings.Repeat("██", len(matrix)+4))
}
