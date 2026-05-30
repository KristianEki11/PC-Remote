//go:build windows

package windows

import (
	"fmt"
	"runtime"
	"strings"
	"syscall"
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
	eRender             = 0
	eCapture            = 1
	eMultimedia         = 1
	CLSCTX_ALL          = 0x17
	DEVICE_STATE_ACTIVE = 0x00000001
	STGM_READ           = 0x00000000
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
// RealAPI struct & initialization
// ──────────────────────────────────────────────────────────

type RealAPI struct{}

func init() {
	API = RealAPI{}
}

// ──────────────────────────────────────────────────────────
// Package-level functions delegating to API
// ──────────────────────────────────────────────────────────

func SetVolume(level float64) error                                  { return API.SetVolume(level) }
func GetVolumeStatus() (level float64, muted bool, err error)        { return API.GetVolumeStatus() }
func SetMute(muted bool) error                                       { return API.SetMute(muted) }
func GetAllChannelVolumes() (map[string]ChannelStatus, error)        { return API.GetAllChannelVolumes() }
func SetChannelVolume(channelName string, level float64) error       { return API.SetChannelVolume(channelName, level) }
func SetChannelMute(channelName string, muted bool) error            { return API.SetChannelMute(channelName, muted) }
func GetDevices() ([]DeviceStatus, error)                            { return API.GetDevices() }
func SetDeviceVolume(id string, level float64) error                 { return API.SetDeviceVolume(id, level) }
func SetDeviceMute(id string, muted bool) error                     { return API.SetDeviceMute(id, muted) }
func SendMediaKey(action string) error                               { return API.SendMediaKey(action) }
func GetMediaStatus() (map[string]any, error)                         { return API.GetMediaStatus() }
func LockWorkstation() error                                         { return API.LockWorkstation() }
func OpenBrowser(url string) error                                   { return API.OpenBrowser(url) }
func ScheduleShutdown(delaySeconds int) error                        { return API.ScheduleShutdown(delaySeconds) }
func CancelShutdown() error                                          { return API.CancelShutdown() }
func Sleep() error                                                   { return API.Sleep() }
func Restart() error                                                 { return API.Restart() }
func DebugListAllDevices() ([]string, error)                         { return API.DebugListAllDevices() }
