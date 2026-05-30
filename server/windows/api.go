//go:build windows

package windows

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
)

// ──────────────────────────────────────────────────────────
// COM GUIDs
// ──────────────────────────────────────────────────────────

var (
	CLSID_MMDeviceEnumerator = ole.NewGUID("{BCDE0395-E52F-467C-8E3D-C4579291692E}")
	IID_IMMDeviceEnumerator  = ole.NewGUID("{A95664D2-9614-4F35-A746-DE8DB63617E6}")
	IID_IAudioEndpointVolume = ole.NewGUID("{5CDF2C82-841E-4546-9722-0CF74078229A}")
	IID_IPropertyStore       = ole.NewGUID("{886D8EEB-8CF2-4446-8D02-CDBA1DBDCF99}")
)

// PKEY_Device_FriendlyName property key for reading device display names.
// {A45C254E-DF1C-4EFD-8020-67D146A850E0}, pid 14
var PKEY_Device_FriendlyName = propertyKey{
	fmtid: ole.NewGUID("{A45C254E-DF1C-4EFD-8020-67D146A850E0}"),
	pid:   14,
}

type propertyKey struct {
	fmtid *ole.GUID
	pid   uint32
}

// Raw PROPERTYKEY as laid out in memory for COM calls (16-byte GUID + 4-byte pid).
type rawPKEY struct {
	fmtid [16]byte
	pid   uint32
}

func pkeyToRaw(pk propertyKey) rawPKEY {
	var raw rawPKEY
	// Copy GUID bytes: Data1(4) + Data2(2) + Data3(2) + Data4(8) = 16
	*(*uint32)(unsafe.Pointer(&raw.fmtid[0])) = pk.fmtid.Data1
	*(*uint16)(unsafe.Pointer(&raw.fmtid[4])) = pk.fmtid.Data2
	*(*uint16)(unsafe.Pointer(&raw.fmtid[6])) = pk.fmtid.Data3
	copy(raw.fmtid[8:], pk.fmtid.Data4[:])
	raw.pid = pk.pid
	return raw
}

// ──────────────────────────────────────────────────────────
// Constants
// ──────────────────────────────────────────────────────────

const (
	eRender            = 0
	eCapture           = 1
	eMultimedia        = 1
	CLSCTX_ALL         = 0x17
	DEVICE_STATE_ACTIVE = 0x00000001
	STGM_READ          = 0x00000000
)

// IAudioEndpointVolume vtable offsets (inherits IUnknown[0-2])
const (
	vtSetMasterVolumeLevelScalar = 7
	vtGetMasterVolumeLevelScalar = 9
	vtSetMute                    = 14
	vtGetMute                    = 15
)

// Media virtual-key codes
const (
	VK_MEDIA_NEXT_TRACK = 0xB0
	VK_MEDIA_PREV_TRACK = 0xB1
	VK_MEDIA_PLAY_PAUSE = 0xB3
	INPUT_KEYBOARD      = 1
	KEYEVENTF_KEYUP     = 0x0002
)

// sonarChannels maps logical channel names to partial friendly-name matches.
// Order matters: first match wins.
var sonarChannels = map[string]string{
	"gaming": "Sonar - Gaming",
	"chat":   "Sonar - Chat",
	"media":  "Sonar - Media",
	"mic":    "Sonar - Microphone",
}

// preferredDevice is the partial name we look for when the caller asks for
// the "default" volume target. If not found, we fall back to the real
// Windows default audio endpoint.
const preferredDevice = "Sonar - Media"

// ──────────────────────────────────────────────────────────
// COM threading helper
// ──────────────────────────────────────────────────────────

func withCOM(fn func() error) error {
	var result error
	done := make(chan struct{})
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		defer close(done)
		if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
			result = fmt.Errorf("CoInitializeEx: %w", err)
			return
		}
		defer ole.CoUninitialize()
		result = fn()
	}()
	<-done
	return result
}

// ──────────────────────────────────────────────────────────
// Low-level COM helpers
// ──────────────────────────────────────────────────────────

// vtbl returns the vtable array for a given COM interface pointer.
func vtbl(ptr uintptr) *[1024]uintptr {
	// The COM interface pointer points to the vtable pointer.
	vtblPtr := *(*uintptr)(unsafe.Pointer(ptr))
	return (*[1024]uintptr)(unsafe.Pointer(vtblPtr))
}

// createEnumerator creates an IMMDeviceEnumerator COM instance.
func createEnumerator() (uintptr, error) {
	enum, err := ole.CreateInstance(CLSID_MMDeviceEnumerator, IID_IMMDeviceEnumerator)
	if err != nil {
		return 0, fmt.Errorf("CreateInstance(MMDeviceEnumerator): %w", err)
	}
	return uintptr(unsafe.Pointer(enum)), nil
}

// activateEndpointVolume calls IMMDevice::Activate for IAudioEndpointVolume.
func activateEndpointVolume(devicePtr uintptr) (uintptr, error) {
	var epPtr uintptr
	ret, _, _ := syscall.SyscallN(vtbl(devicePtr)[3],
		devicePtr,
		uintptr(unsafe.Pointer(IID_IAudioEndpointVolume)),
		uintptr(CLSCTX_ALL),
		0,
		uintptr(unsafe.Pointer(&epPtr)),
	)
	if ret != 0 {
		return 0, fmt.Errorf("IMMDevice.Activate failed: HRESULT 0x%08X", ret)
	}
	return epPtr, nil
}

// releasePtr calls IUnknown::Release on a raw COM pointer.
func releasePtr(ptr uintptr) {
	if ptr == 0 {
		return
	}
	syscall.SyscallN(vtbl(ptr)[2], ptr) // vtable[2] = Release
}

// getDeviceFriendlyName reads PKEY_Device_FriendlyName from an IMMDevice.
func getDeviceFriendlyName(devicePtr uintptr) string {
	// IMMDevice::OpenPropertyStore(STGM_READ, &propStore)
	var propStorePtr uintptr
	ret, _, _ := syscall.SyscallN(vtbl(devicePtr)[4],
		devicePtr,
		uintptr(STGM_READ),
		uintptr(unsafe.Pointer(&propStorePtr)),
	)
	if ret != 0 || propStorePtr == 0 {
		return ""
	}
	defer releasePtr(propStorePtr)

	// IPropertyStore::GetValue(PROPERTYKEY, PROPVARIANT*)
	raw := pkeyToRaw(PKEY_Device_FriendlyName)

	// PROPVARIANT is 24 bytes on x64 (vt[2] + pad[6] + val[16])
	var propVar [24]byte
	ret, _, _ = syscall.SyscallN(vtbl(propStorePtr)[5],
		propStorePtr,
		uintptr(unsafe.Pointer(&raw)),
		uintptr(unsafe.Pointer(&propVar[0])),
	)
	if ret != 0 {
		return ""
	}

	// vt = VT_LPWSTR (31). The LPWSTR pointer sits at offset 8.
	vt := *(*uint16)(unsafe.Pointer(&propVar[0]))
	if vt != 31 { // VT_LPWSTR
		return ""
	}
	strPtr := *(*uintptr)(unsafe.Pointer(&propVar[8]))
	if strPtr == 0 {
		return ""
	}

	// Read null-terminated UTF-16 string
	name := readUTF16(strPtr)

	// Free the BSTR/CoTaskMem allocated string via PropVariantClear
	pOle32 := syscall.NewLazyDLL("ole32.dll")
	pClear := pOle32.NewProc("PropVariantClear")
	pClear.Call(uintptr(unsafe.Pointer(&propVar[0])))

	return name
}

func readUTF16(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	var chars []uint16
	for i := 0; ; i++ {
		ch := *(*uint16)(unsafe.Pointer(ptr + uintptr(i*2)))
		if ch == 0 {
			break
		}
		chars = append(chars, ch)
		if i > 512 { // safety limit
			break
		}
	}
	return syscall.UTF16ToString(chars)
}

// ──────────────────────────────────────────────────────────
// Device enumeration
// ──────────────────────────────────────────────────────────

// deviceInfo holds a discovered audio device.
type deviceInfo struct {
	ptr  uintptr // raw IMMDevice*, caller must release
	name string
}

// enumRenderDevices returns all active render (output) audio endpoints.
// Caller must releasePtr each device when done.
func enumRenderDevices(enumPtr uintptr) ([]deviceInfo, error) {
	return enumDevices(enumPtr, eRender)
}

// enumCaptureDevices returns all active capture (input) audio endpoints.
func enumCaptureDevices(enumPtr uintptr) ([]deviceInfo, error) {
	return enumDevices(enumPtr, eCapture)
}

func enumDevices(enumPtr uintptr, dataFlow int) ([]deviceInfo, error) {
	// IMMDeviceEnumerator::EnumAudioEndpoints(dataFlow, stateMask, &collection)
	var collPtr uintptr
	ret, _, _ := syscall.SyscallN(vtbl(enumPtr)[3],
		enumPtr,
		uintptr(dataFlow),
		uintptr(DEVICE_STATE_ACTIVE),
		uintptr(unsafe.Pointer(&collPtr)),
	)
	if ret != 0 {
		return nil, fmt.Errorf("EnumAudioEndpoints failed: HRESULT 0x%08X", ret)
	}
	defer releasePtr(collPtr)

	// IMMDeviceCollection::GetCount(&count)
	var count uint32
	ret, _, _ = syscall.SyscallN(vtbl(collPtr)[3],
		collPtr,
		uintptr(unsafe.Pointer(&count)),
	)
	if ret != 0 {
		return nil, fmt.Errorf("GetCount failed: HRESULT 0x%08X", ret)
	}

	devices := make([]deviceInfo, 0, count)
	for i := uint32(0); i < count; i++ {
		// IMMDeviceCollection::Item(i, &device)
		var devPtr uintptr
		ret, _, _ = syscall.SyscallN(vtbl(collPtr)[4],
			collPtr,
			uintptr(i),
			uintptr(unsafe.Pointer(&devPtr)),
		)
		if ret != 0 {
			continue
		}
		name := getDeviceFriendlyName(devPtr)
		devices = append(devices, deviceInfo{ptr: devPtr, name: name})
	}
	return devices, nil
}

// findDeviceByName searches devices for one whose friendly name contains partialName.
func findDeviceByName(devices []deviceInfo, partialName string) (uintptr, bool) {
	lower := strings.ToLower(partialName)
	for _, d := range devices {
		if strings.Contains(strings.ToLower(d.name), lower) {
			return d.ptr, true
		}
	}
	return 0, false
}

// releaseDevices releases all device pointers except the one at keepPtr.
func releaseDevices(devices []deviceInfo, keepPtr uintptr) {
	for _, d := range devices {
		if d.ptr != keepPtr {
			releasePtr(d.ptr)
		}
	}
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
// RealAPI Implementation of APIInterface
// ──────────────────────────────────────────────────────────

type RealAPI struct{}

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

var (
	kernel32              = syscall.NewLazyDLL("kernel32.dll")
	pProcessIdToSessionId = kernel32.NewProc("ProcessIdToSessionId")
)

// runInUserSession executes a command in the active interactive user session.
// If running in Session 0 (as a Windows service), it registers a temporary
// scheduled task to bypass Session 0 isolation. Otherwise, it runs the command directly.
func runInUserSession(command string) error {
	var sessionID uint32
	pid := uint32(os.Getpid())
	ret, _, _ := pProcessIdToSessionId.Call(uintptr(pid), uintptr(unsafe.Pointer(&sessionID)))
	
	// If we are already running in an interactive user session (Session ID > 0),
	// we do not need to use schtasks to bypass Session 0 isolation.
	if ret != 0 && sessionID > 0 {
		slog.Info("Running command directly (interactive session)", "session_id", sessionID, "command", command)
		cmd := exec.Command("cmd.exe", "/c", command)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("direct command execution failed: %w — %s", err, string(out))
		}
		return nil
	}

	slog.Info("Running command via schtasks (Session 0 service)", "command", command)
	// Use a unique task name to avoid conflicts
	taskName := fmt.Sprintf("PCRemoteTask_%d", syscall.Getpid())

	// Register the task to run once immediately (interactive = all logged on users)
	register := exec.Command("schtasks", "/create", "/tn", taskName,
		"/tr", command,
		"/sc", "ONCE",
		"/st", "00:00",
		"/f",
		"/ru", "INTERACTIVE",
	)
	if out, err := register.CombinedOutput(); err != nil {
		return fmt.Errorf("schtasks /create failed: %w — %s", err, string(out))
	}

	// Run the task immediately
	run := exec.Command("schtasks", "/run", "/tn", taskName)
	if out, err := run.CombinedOutput(); err != nil {
		// Try to clean up even if run failed
		exec.Command("schtasks", "/delete", "/tn", taskName, "/f").Run() //nolint
		return fmt.Errorf("schtasks /run failed: %w — %s", err, string(out))
	}

	// Delete the task in a background goroutine after a short sleep to allow the process to finish
	go func(tn string) {
		time.Sleep(1 * time.Second)
		exec.Command("schtasks", "/delete", "/tn", tn, "/f").Run()
	}(taskName)
	return nil
}

func (RealAPI) SendMediaKey(action string) error {
	action = strings.ToLower(action)
	if action != "play_pause" && action != "next" && action != "prev" {
		return fmt.Errorf("unknown media action: %q", action)
	}

	exePath := filepath.Join(os.Getenv("PROGRAMFILES"), "PCRemote", "sendkey.exe")
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		exePath = filepath.Join("d:\\remote-pc\\server", "sendkey.exe")
	}

	cmd := fmt.Sprintf(`"%s" %s`, exePath, action)
	return runInUserSession(cmd)
}

func (RealAPI) LockWorkstation() error {
	return runInUserSession(`rundll32.exe user32.dll,LockWorkStation`)
}

func (RealAPI) OpenBrowser(url string) error {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return errors.New("url must start with http:// or https://")
	}
	// Open via default browser in user session
	return runInUserSession(fmt.Sprintf(`rundll32 url.dll,FileProtocolHandler %s`, url))
}

func (RealAPI) ScheduleShutdown(delaySeconds int) error {
	delaySec := strconv.Itoa(delaySeconds)
	cmd := exec.Command("shutdown", "/s", "/t", delaySec)
	return cmd.Run()
}

func (RealAPI) CancelShutdown() error {
	cmd := exec.Command("shutdown", "/a")
	return cmd.Run()
}

var (
	powrprof               = syscall.NewLazyDLL("powrprof.dll")
	pSetSuspendState       = powrprof.NewProc("SetSuspendState")
	advapi32               = syscall.NewLazyDLL("advapi32.dll")
	pOpenProcessToken      = advapi32.NewProc("OpenProcessToken")
	pLookupPrivilegeValue  = advapi32.NewProc("LookupPrivilegeValueW")
	pAdjustTokenPrivileges = advapi32.NewProc("AdjustTokenPrivileges")
)

func enableShutdownPrivilege() error {
	const (
		TOKEN_ADJUST_PRIVILEGES = 0x0020
		TOKEN_QUERY             = 0x0008
		SE_PRIVILEGE_ENABLED    = 0x00000002
	)

	type LUID struct {
		LowPart  uint32
		HighPart int32
	}

	type LUID_AND_ATTRIBUTES struct {
		Luid       LUID
		Attributes uint32
	}

	type TOKEN_PRIVILEGES struct {
		PrivilegeCount uint32
		Privileges     [1]LUID_AND_ATTRIBUTES
	}

	currentProcess := syscall.Handle(^uintptr(0))
	var token syscall.Handle
	ret, _, err := pOpenProcessToken.Call(
		uintptr(currentProcess),
		TOKEN_ADJUST_PRIVILEGES|TOKEN_QUERY,
		uintptr(unsafe.Pointer(&token)),
	)
	if ret == 0 {
		return fmt.Errorf("OpenProcessToken failed: %w", err)
	}
	defer syscall.CloseHandle(token)

	var luid LUID
	privilegeName := syscall.StringToUTF16Ptr("SeShutdownPrivilege")
	ret, _, err = pLookupPrivilegeValue.Call(
		0,
		uintptr(unsafe.Pointer(privilegeName)),
		uintptr(unsafe.Pointer(&luid)),
	)
	if ret == 0 {
		return fmt.Errorf("LookupPrivilegeValue failed: %w", err)
	}

	tp := TOKEN_PRIVILEGES{
		PrivilegeCount: 1,
		Privileges: [1]LUID_AND_ATTRIBUTES{
			{
				Luid:       luid,
				Attributes: SE_PRIVILEGE_ENABLED,
			},
		},
	}

	ret, _, err = pAdjustTokenPrivileges.Call(
		uintptr(token),
		0,
		uintptr(unsafe.Pointer(&tp)),
		0,
		0,
		0,
	)
	if ret == 0 {
		return fmt.Errorf("AdjustTokenPrivileges failed: %w", err)
	}

	return nil
}

func (RealAPI) Sleep() error {
	// 1. Try native SetSuspendState (works for traditional S3 sleep systems)
	if err := enableShutdownPrivilege(); err != nil {
		slog.Error("Failed to enable shutdown privilege", "error", err)
	}
	nativeRet, _, nativeErr := pSetSuspendState.Call(0, 0, 0)
	slog.Info("Native SetSuspendState call completed", "ret", nativeRet, "err", nativeErr)

	// 2. Support for Modern Standby (S0) systems by turning off the display.
	// We write a temporary, self-deleting PowerShell script in C:\Users\Public
	// and run it via a temporary scheduled task in the user session.
	psPath := `C:\Users\Public\sleep_temp.ps1`
	psContent := `$code = '[DllImport("user32.dll")] public static extern int SendMessage(int h, int m, int w, int l);'
$type = Add-Type -MemberDefinition $code -Name Win32 -PassThru
$type::SendMessage(-1, 0x0112, 0xF170, 2)
Remove-Item $PSCommandPath -Force
`
	if err := os.WriteFile(psPath, []byte(psContent), 0666); err == nil {
		cmdStr := fmt.Sprintf(`powershell.exe -NoProfile -NonInteractive -ExecutionPolicy Bypass -File "%s"`, psPath)
		if runErr := runInUserSession(cmdStr); runErr != nil {
			slog.Error("Failed to run Modern Standby sleep helper in user session", "error", runErr)
			// Clean up file if task scheduling failed
			if _, statErr := os.Stat(psPath); statErr == nil {
				os.Remove(psPath)
			}
		} else {
			slog.Info("Modern Standby sleep helper executed successfully in user session")
		}
	} else {
		slog.Error("Failed to write temporary sleep helper script to C:\\Users\\Public", "error", err)
	}

	return nil
}

func (RealAPI) Restart() error {
	cmd := exec.Command("shutdown", "/r", "/t", "5")
	return cmd.Run()
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

func init() {
	API = RealAPI{}
}

// ──────────────────────────────────────────────────────────
// Package-level functions delegating to API
// ──────────────────────────────────────────────────────────

func SetVolume(level float64) error {
	return API.SetVolume(level)
}

func GetVolumeStatus() (level float64, muted bool, err error) {
	return API.GetVolumeStatus()
}

func SetMute(muted bool) error {
	return API.SetMute(muted)
}

func GetAllChannelVolumes() (map[string]ChannelStatus, error) {
	return API.GetAllChannelVolumes()
}

func SetChannelVolume(channelName string, level float64) error {
	return API.SetChannelVolume(channelName, level)
}

func SetChannelMute(channelName string, muted bool) error {
	return API.SetChannelMute(channelName, muted)
}

func GetDevices() ([]DeviceStatus, error) {
	return API.GetDevices()
}

func SetDeviceVolume(id string, level float64) error {
	return API.SetDeviceVolume(id, level)
}

func SetDeviceMute(id string, muted bool) error {
	return API.SetDeviceMute(id, muted)
}

func SendMediaKey(action string) error {
	return API.SendMediaKey(action)
}

func LockWorkstation() error {
	return API.LockWorkstation()
}

func OpenBrowser(url string) error {
	return API.OpenBrowser(url)
}

func ScheduleShutdown(delaySeconds int) error {
	return API.ScheduleShutdown(delaySeconds)
}

func CancelShutdown() error {
	return API.CancelShutdown()
}

func Sleep() error {
	return API.Sleep()
}

func Restart() error {
	return API.Restart()
}

func DebugListAllDevices() ([]string, error) {
	return API.DebugListAllDevices()
}

// ──────────────────────────────────────────────────────────
// PUBLIC API: Media Keys (SendInput) Helpers
// ──────────────────────────────────────────────────────────

var (
	user32     = syscall.NewLazyDLL("user32.dll")
	pSendInput = user32.NewProc("SendInput")
)

type kbInput struct {
	Type  uint32
	pad1  [4]byte
	Vk    uint16
	Scan  uint16
	Flags uint32
	Time  uint32
	Extra uintptr
	pad2  [4]byte
}

var pLockWorkStation = user32.NewProc("LockWorkStation")
