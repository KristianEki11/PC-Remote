package main

import (
	"os"
	"strings"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		return
	}
	action := strings.ToLower(os.Args[1])
	var vk byte
	switch action {
	case "play_pause":
		vk = 0xB3
	case "next":
		vk = 0xB0
	case "prev":
		vk = 0xB1
	default:
		return
	}

	user32 := syscall.NewLazyDLL("user32.dll")
	pKeybdEvent := user32.NewProc("keybd_event")
	pMapVirtualKey := user32.NewProc("MapVirtualKeyW")

	// Get actual hardware scan code
	scanCode, _, _ := pMapVirtualKey.Call(uintptr(vk), 0)

	// Key down (KEYEVENTF_EXTENDEDKEY = 0x0001)
	pKeybdEvent.Call(uintptr(vk), scanCode, 1, 0)
	// Key up (KEYEVENTF_KEYUP = 0x0002 | KEYEVENTF_EXTENDEDKEY = 0x0001)
	pKeybdEvent.Call(uintptr(vk), scanCode, 2|1, 0)
}
