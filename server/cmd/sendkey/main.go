package main

import (
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// INPUT type constants
const (
	inputKeyboard = 1

	keyeventfExtendedKey = 0x0001
	keyeventfKeyUp       = 0x0002
)

// Virtual key codes for media keys
const (
	vkMediaPlayPause  = 0xB3
	vkMediaNextTrack  = 0xB0
	vkMediaPrevTrack  = 0xB1
)

// KEYBDINPUT structure mirrors Win32 KEYBDINPUT
// https://learn.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-keybdinput
type KEYBDINPUT struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

// INPUT structure mirrors Win32 INPUT (keyboard variant)
// Total size must be 28 bytes on 64-bit: type(4) + padding(4) + KEYBDINPUT(20)
type INPUT struct {
	Type uint32
	_    [4]byte // alignment padding
	Ki   KEYBDINPUT
}

var (
	user32        = syscall.NewLazyDLL("user32.dll")
	pSendInput    = user32.NewProc("SendInput")
	pMapVirtualKey = user32.NewProc("MapVirtualKeyW")
)

func sendKey(vk uint16) {
	// Resolve hardware scan code from virtual key code
	scanCode, _, _ := pMapVirtualKey.Call(uintptr(vk), 0)

	inputs := [2]INPUT{
		// Key Down
		{
			Type: inputKeyboard,
			Ki: KEYBDINPUT{
				WVk:     vk,
				WScan:   uint16(scanCode),
				DwFlags: keyeventfExtendedKey,
			},
		},
		// Key Up
		{
			Type: inputKeyboard,
			Ki: KEYBDINPUT{
				WVk:     vk,
				WScan:   uint16(scanCode),
				DwFlags: keyeventfExtendedKey | keyeventfKeyUp,
			},
		},
	}

	pSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(unsafe.Sizeof(inputs[0])),
	)
}

func main() {
	if len(os.Args) < 2 {
		return
	}
	action := strings.ToLower(os.Args[1])

	switch action {
	case "play_pause":
		sendKey(vkMediaPlayPause)
	case "next":
		sendKey(vkMediaNextTrack)
	case "prev":
		sendKey(vkMediaPrevTrack)
	}
}
