//go:build windows

package windows

import (
	"fmt"
	"log/slog"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var (
	brightnessMutex sync.Mutex
	cachedSliderVal float64 = -1
)

// GetBrightness returns a combined brightness value in the range 0.0 - 100.0:
// - Range 50.0 - 100.0: Hardware brightness = (slider - 50) * 2 (Overlay 0%)
// - Range 0.0 - 50.0: Hardware brightness = 0, Overlay opacity = (50 - slider) * 2%
func (RealAPI) GetBrightness() (float64, error) {
	brightnessMutex.Lock()
	defer brightnessMutex.Unlock()

	// If overlay is currently dimming (< 50%), return cached slider value if valid
	overlayMutex.Lock()
	running := isOverlayRunning
	alpha := currentAlpha
	overlayMutex.Unlock()

	if running && alpha > 0 {
		if cachedSliderVal >= 0 && cachedSliderVal < 50.0 {
			return cachedSliderVal, nil
		}
		// Derive from alpha: alpha = byte(math.Round((50.0 - slider) * 2.0 * 255.0 / 100.0))
		pct := float64(alpha) / 255.0
		slider := 50.0 - (pct * 50.0)
		return math.Round(slider), nil
	}

	hw, err := getHardwareBrightness()
	if err != nil {
		slog.Warn("Failed to read hardware brightness, using cache or fallback", "error", err)
		if cachedSliderVal >= 0 {
			return cachedSliderVal, nil
		}
		return 50.0, nil
	}

	// Map hw (0..100) -> slider (50..100)
	slider := 50.0 + (hw / 2.0)
	cachedSliderVal = slider
	return math.Round(slider), nil
}

// SetBrightness sets the combined brightness slider (0.0 - 100.0):
// - 50.0 - 100.0: Hardware backlight is (level - 50) * 2, overlay closed
// - < 50.0: Hardware backlight is 0, overlay opacity is (50 - level) * 2%
func (RealAPI) SetBrightness(level float64) error {
	if level < 0.0 {
		level = 0.0
	} else if level > 100.0 {
		level = 100.0
	}

	brightnessMutex.Lock()
	defer brightnessMutex.Unlock()

	cachedSliderVal = level

	if level >= 50.0 {
		// Close overlay
		CloseOverlay()

		// Hardware scale 0 - 100
		hw := (level - 50.0) * 2.0
		hwInt := int(math.Round(hw))
		return setHardwareBrightness(hwInt)
	}

	// Under 50%: Hardware backlight set to 0
	_ = setHardwareBrightness(0)

	// Calculate overlay alpha (0 - 255)
	// At 49%: opacity 2% -> alpha ~ 5
	// At 0%: opacity 100% -> alpha 255
	overlayPercent := (50.0 - level) * 2.0 // 0 to 100
	alphaVal := byte(math.Round(overlayPercent * 255.0 / 100.0))

	// If level == 0%, make it full black non-click-through (like display off)
	clickThrough := level > 0.0

	SetOverlayOpacity(alphaVal, clickThrough)
	return nil
}

// getHardwareBrightness reads the current hardware brightness (0-100) via WMI.
func getHardwareBrightness() (float64, error) {
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command",
		"(Get-WmiObject -Namespace root\\wmi -Class WmiMonitorBrightness).CurrentBrightness")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	str := strings.TrimSpace(string(out))
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid brightness output '%s': %w", str, err)
	}

	return val, nil
}

// setHardwareBrightness writes the hardware brightness (0-100) via WMI.
func setHardwareBrightness(level int) error {
	if level < 0 {
		level = 0
	} else if level > 100 {
		level = 100
	}

	psScript := fmt.Sprintf("(Get-WmiObject -Namespace root\\wmi -Class WmiMonitorBrightnessMethods).WmiSetBrightness(0, %d)", level)
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	return cmd.Run()
}
