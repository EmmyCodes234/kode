//go:build !windows

package voice

import (
	"context"
	"errors"
)

func (vb *VoiceBridge) recordWindows(ctx context.Context, outputPath string, duration int) error {
	return errors.New("native Windows recording (MCI DLL) is not supported on this platform")
}
