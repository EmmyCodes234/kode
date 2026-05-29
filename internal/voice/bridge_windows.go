//go:build windows

package voice

import (
	"context"
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

var (
	winmm          = syscall.NewLazyDLL("winmm.dll")
	mciSendStringA = winmm.NewProc("mciSendStringA")
)

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
