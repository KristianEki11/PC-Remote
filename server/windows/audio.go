//go:build windows

package windows

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"syscall"
	"unsafe"
)

// ──────────────────────────────────────────────────────────
// Audio: Master volume, Sonar channels, and device management
// ──────────────────────────────────────────────────────────

// preferredDevice is the partial name we look for when the caller asks for
// the "default" volume target. If not found, we fall back to the real
// Windows default audio endpoint.
const preferredDevice = "Sonar - Media"

// sonarChannels maps logical channel names to partial friendly-name matches.
// Order matters: first match wins.
var sonarChannels = map[string]string{
	"gaming": "Sonar - Gaming",
	"chat":   "Sonar - Chat",
	"media":  "Sonar - Media",
	"mic":    "Sonar - Microphone",
}

// ──────────────────────────────────────────────────────────
// getTargetEndpointVolume: Sonar-aware device resolution
// ──────────────────────────────────────────────────────────

// getTargetEndpointVolume finds the preferred audio device ("Sonar Media")
// and activates IAudioEndpointVolume on it. Falls back to the Windows
// default endpoint if Sonar is not installed.
//
// Returns: enumerator, device, endpoint pointers (all must be released).
func getTargetEndpointVolume() (enumPtr, devPtr, epPtr uintptr, err error) {
	enumPtr, err = createEnumerator()
	if err != nil {
		return 0, 0, 0, err
	}

	// Try to find Sonar Media by enumerating all render devices
	devices, eErr := enumRenderDevices(enumPtr)
	if eErr == nil {
		if ptr, ok := findDeviceByName(devices, preferredDevice); ok {
			slog.Debug("Audio target resolved", "device", preferredDevice)
			releaseDevices(devices, ptr)
			devPtr = ptr
			epPtr, err = activateEndpointVolume(devPtr)
			if err != nil {
				releasePtr(devPtr)
				releasePtr(enumPtr)
				return 0, 0, 0, err
			}
			return enumPtr, devPtr, epPtr, nil
		}
		// Sonar not found, release all enumerated devices
		releaseDevices(devices, 0)
	}

	// Fallback: Windows default audio endpoint
	slog.Debug("Sonar Media not found, using default endpoint")
	var defPtr uintptr
	ret, _, _ := syscall.SyscallN(vtbl(enumPtr)[4],
		enumPtr,
		uintptr(eRender),
		uintptr(eMultimedia),
		uintptr(unsafe.Pointer(&defPtr)),
	)
	if ret != 0 {
		releasePtr(enumPtr)
		return 0, 0, 0, fmt.Errorf("GetDefaultAudioEndpoint failed: HRESULT 0x%08X", ret)
	}
	devPtr = defPtr

	epPtr, err = activateEndpointVolume(devPtr)
	if err != nil {
		releasePtr(devPtr)
		releasePtr(enumPtr)
		return 0, 0, 0, err
	}
	return enumPtr, devPtr, epPtr, nil
}

func releaseTriple(enumPtr, devPtr, epPtr uintptr) {
	releasePtr(epPtr)
	releasePtr(devPtr)
	releasePtr(enumPtr)
}

// ──────────────────────────────────────────────────────────
// RealAPI — Audio methods
// ──────────────────────────────────────────────────────────

func (RealAPI) SetVolume(level float64) error {
	if level < 0.0 || level > 1.0 {
		return errors.New("volume level must be between 0.0 and 1.0")
	}
	return withCOM(func() error {
		enumP, devP, epP, err := getTargetEndpointVolume()
		if err != nil {
			return err
		}
		defer releaseTriple(enumP, devP, epP)

		fLevel := float32(level)
		ret, _, _ := syscall.SyscallN(vtbl(epP)[vtSetMasterVolumeLevelScalar],
			epP,
			uintptr(*(*uint32)(unsafe.Pointer(&fLevel))),
			0,
		)
		if ret != 0 {
			return fmt.Errorf("SetMasterVolumeLevelScalar failed: HRESULT 0x%08X", ret)
		}
		return nil
	})
}

func (RealAPI) GetVolumeStatus() (level float64, muted bool, err error) {
	err = withCOM(func() error {
		enumP, devP, epP, cErr := getTargetEndpointVolume()
		if cErr != nil {
			return cErr
		}
		defer releaseTriple(enumP, devP, epP)

		var fLevel float32
		ret, _, _ := syscall.SyscallN(vtbl(epP)[vtGetMasterVolumeLevelScalar],
			epP, uintptr(unsafe.Pointer(&fLevel)),
		)
		if ret != 0 {
			return fmt.Errorf("GetMasterVolumeLevelScalar failed: HRESULT 0x%08X", ret)
		}
		level = float64(fLevel)

		var bMute int32
		ret, _, _ = syscall.SyscallN(vtbl(epP)[vtGetMute],
			epP, uintptr(unsafe.Pointer(&bMute)),
		)
		if ret != 0 {
			return fmt.Errorf("GetMute failed: HRESULT 0x%08X", ret)
		}
		muted = bMute != 0
		return nil
	})
	return
}

func (RealAPI) SetMute(muted bool) error {
	return withCOM(func() error {
		enumP, devP, epP, err := getTargetEndpointVolume()
		if err != nil {
			return err
		}
		defer releaseTriple(enumP, devP, epP)

		var bMute uintptr
		if muted {
			bMute = 1
		}
		ret, _, _ := syscall.SyscallN(vtbl(epP)[vtSetMute], epP, bMute, 0)
		if ret != 0 {
			return fmt.Errorf("SetMute failed: HRESULT 0x%08X", ret)
		}
		return nil
	})
}

func (RealAPI) GetAllChannelVolumes() (map[string]ChannelStatus, error) {
	result := make(map[string]ChannelStatus)
	err := withCOM(func() error {
		enumPtr, cErr := createEnumerator()
		if cErr != nil {
			return cErr
		}
		defer releasePtr(enumPtr)

		// Render devices (gaming, chat, media)
		renderDevs, rErr := enumRenderDevices(enumPtr)
		if rErr != nil {
			return rErr
		}
		defer releaseDevices(renderDevs, 0)

		// Capture devices (mic)
		captureDevs, cErr2 := enumCaptureDevices(enumPtr)
		if cErr2 != nil {
			return cErr2
		}
		defer releaseDevices(captureDevs, 0)

		allDevs := append(renderDevs, captureDevs...)

		for chKey, partialName := range sonarChannels {
			devPtr, ok := findDeviceByName(allDevs, partialName)
			if !ok {
				continue
			}
			epPtr, aErr := activateEndpointVolume(devPtr)
			if aErr != nil {
				continue
			}

			var fLevel float32
			syscall.SyscallN(vtbl(epPtr)[vtGetMasterVolumeLevelScalar],
				epPtr, uintptr(unsafe.Pointer(&fLevel)),
			)

			var bMute int32
			syscall.SyscallN(vtbl(epPtr)[vtGetMute],
				epPtr, uintptr(unsafe.Pointer(&bMute)),
			)

			releasePtr(epPtr)

			result[chKey] = ChannelStatus{
				Name:  partialName,
				Level: math.Round(float64(fLevel)*100) / 100,
				Muted: bMute != 0,
			}
		}
		return nil
	})
	return result, err
}

func (RealAPI) SetChannelVolume(channelName string, level float64) error {
	if level < 0.0 || level > 1.0 {
		return errors.New("volume level must be between 0.0 and 1.0")
	}
	partialName, ok := sonarChannels[strings.ToLower(channelName)]
	if !ok {
		return fmt.Errorf("unknown channel %q (valid: gaming, chat, media, mic)", channelName)
	}

	return withCOM(func() error {
		enumPtr, cErr := createEnumerator()
		if cErr != nil {
			return cErr
		}
		defer releasePtr(enumPtr)

		// Search both render and capture (mic is capture)
		renderDevs, _ := enumRenderDevices(enumPtr)
		captureDevs, _ := enumCaptureDevices(enumPtr)
		allDevs := append(renderDevs, captureDevs...)
		defer releaseDevices(allDevs, 0)

		devPtr, found := findDeviceByName(allDevs, partialName)
		if !found {
			return fmt.Errorf("Sonar channel %q (%s) not found", channelName, partialName)
		}

		epPtr, aErr := activateEndpointVolume(devPtr)
		if aErr != nil {
			return aErr
		}
		defer releasePtr(epPtr)

		fLevel := float32(level)
		ret, _, _ := syscall.SyscallN(vtbl(epPtr)[vtSetMasterVolumeLevelScalar],
			epPtr,
			uintptr(*(*uint32)(unsafe.Pointer(&fLevel))),
			0,
		)
		if ret != 0 {
			return fmt.Errorf("SetMasterVolumeLevelScalar failed: HRESULT 0x%08X", ret)
		}
		return nil
	})
}

func (RealAPI) SetChannelMute(channelName string, muted bool) error {
	partialName, ok := sonarChannels[strings.ToLower(channelName)]
	if !ok {
		return fmt.Errorf("unknown channel %q (valid: gaming, chat, media, mic)", channelName)
	}

	return withCOM(func() error {
		enumPtr, cErr := createEnumerator()
		if cErr != nil {
			return cErr
		}
		defer releasePtr(enumPtr)

		renderDevs, _ := enumRenderDevices(enumPtr)
		captureDevs, _ := enumCaptureDevices(enumPtr)
		allDevs := append(renderDevs, captureDevs...)
		defer releaseDevices(allDevs, 0)

		devPtr, found := findDeviceByName(allDevs, partialName)
		if !found {
			return fmt.Errorf("Sonar channel %q (%s) not found", channelName, partialName)
		}

		epPtr, aErr := activateEndpointVolume(devPtr)
		if aErr != nil {
			return aErr
		}
		defer releasePtr(epPtr)

		var bMute uintptr
		if muted {
			bMute = 1
		}
		ret, _, _ := syscall.SyscallN(vtbl(epPtr)[vtSetMute], epPtr, bMute, 0)
		if ret != 0 {
			return fmt.Errorf("SetMute failed: HRESULT 0x%08X", ret)
		}
		return nil
	})
}

func getDeviceID(devicePtr uintptr) string {
	var strPtr uintptr
	ret, _, _ := syscall.SyscallN(vtbl(devicePtr)[5], // IMMDevice::GetId
		devicePtr,
		uintptr(unsafe.Pointer(&strPtr)),
	)
	if ret != 0 || strPtr == 0 {
		return ""
	}
	defer func() {
		pOle32 := syscall.NewLazyDLL("ole32.dll")
		pFree := pOle32.NewProc("CoTaskMemFree")
		pFree.Call(strPtr)
	}()
	return readUTF16(strPtr)
}

func (RealAPI) GetDevices() ([]DeviceStatus, error) {
	var statuses []DeviceStatus
	err := withCOM(func() error {
		enumPtr, cErr := createEnumerator()
		if cErr != nil {
			return cErr
		}
		defer releasePtr(enumPtr)

		renderDevs, rErr := enumRenderDevices(enumPtr)
		if rErr != nil {
			return rErr
		}
		defer releaseDevices(renderDevs, 0)

		for _, d := range renderDevs {
			id := getDeviceID(d.ptr)
			if id == "" {
				continue
			}

			epPtr, aErr := activateEndpointVolume(d.ptr)
			if aErr != nil {
				continue
			}

			var fLevel float32
			syscall.SyscallN(vtbl(epPtr)[vtGetMasterVolumeLevelScalar],
				epPtr, uintptr(unsafe.Pointer(&fLevel)),
			)

			var bMute int32
			syscall.SyscallN(vtbl(epPtr)[vtGetMute],
				epPtr, uintptr(unsafe.Pointer(&bMute)),
			)

			releasePtr(epPtr)

			statuses = append(statuses, DeviceStatus{
				ID:     id,
				Name:   d.name,
				Volume: float64(fLevel * 100.0), // Convert scalar to 0-100
				Muted:  bMute != 0,
			})
		}
		return nil
	})
	return statuses, err
}

func (RealAPI) SetDeviceVolume(id string, level float64) error {
	return withCOM(func() error {
		enumPtr, cErr := createEnumerator()
		if cErr != nil {
			return cErr
		}
		defer releasePtr(enumPtr)

		renderDevs, rErr := enumRenderDevices(enumPtr)
		if rErr != nil {
			return rErr
		}
		defer releaseDevices(renderDevs, 0)

		var targetPtr uintptr
		for _, d := range renderDevs {
			dId := getDeviceID(d.ptr)
			if dId == id {
				targetPtr = d.ptr
				break
			}
		}

		if targetPtr == 0 {
			return fmt.Errorf("device not found: %s", id)
		}

		epPtr, aErr := activateEndpointVolume(targetPtr)
		if aErr != nil {
			return aErr
		}
		defer releasePtr(epPtr)

		// Set volume
		ret, _, _ := syscall.SyscallN(vtbl(epPtr)[vtSetMasterVolumeLevelScalar],
			epPtr, uintptr(math.Float32bits(float32(level))), 0,
		)
		if ret != 0 {
			return fmt.Errorf("SetMasterVolumeLevelScalar failed: HRESULT 0x%08X", ret)
		}
		return nil
	})
}

func (RealAPI) SetDeviceMute(id string, muted bool) error {
	return withCOM(func() error {
		enumPtr, cErr := createEnumerator()
		if cErr != nil {
			return cErr
		}
		defer releasePtr(enumPtr)

		renderDevs, rErr := enumRenderDevices(enumPtr)
		if rErr != nil {
			return rErr
		}
		defer releaseDevices(renderDevs, 0)

		var targetPtr uintptr
		for _, d := range renderDevs {
			dId := getDeviceID(d.ptr)
			if dId == id {
				targetPtr = d.ptr
				break
			}
		}

		if targetPtr == 0 {
			return fmt.Errorf("device not found: %s", id)
		}

		epPtr, aErr := activateEndpointVolume(targetPtr)
		if aErr != nil {
			return aErr
		}
		defer releasePtr(epPtr)

		var muteVal int32
		if muted {
			muteVal = 1
		}
		ret, _, _ := syscall.SyscallN(vtbl(epPtr)[vtSetMute],
			epPtr, uintptr(muteVal), 0,
		)
		if ret != 0 {
			return fmt.Errorf("SetMute failed: HRESULT 0x%08X", ret)
		}
		return nil
	})
}

func (RealAPI) DebugListAllDevices() ([]string, error) {
	var names []string
	err := withCOM(func() error {
		enumPtr, cErr := createEnumerator()
		if cErr != nil {
			return cErr
		}
		defer releasePtr(enumPtr)

		renderDevs, rErr := enumRenderDevices(enumPtr)
		if rErr != nil {
			return rErr
		}
		defer releaseDevices(renderDevs, 0)

		captureDevs, cErr2 := enumCaptureDevices(enumPtr)
		if cErr2 != nil {
			return cErr2
		}
		defer releaseDevices(captureDevs, 0)

		for _, d := range renderDevs {
			names = append(names, "[RENDER] " + d.name)
		}
		for _, d := range captureDevs {
			names = append(names, "[CAPTURE] " + d.name)
		}
		return nil
	})
	return names, err
}
