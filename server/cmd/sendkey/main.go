package main

import (
	"os"
	"strings"
	"syscall"
	"unsafe"
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

func main() {
	if len(os.Args) < 2 {
		return
	}
	action := strings.ToLower(os.Args[1])
	var vk uint16
	var scan uint16
	switch action {
	case "play_pause":
		vk = 0xB3
		scan = 0x22
	case "next":
		vk = 0xB0
		scan = 0x19
	case "prev":
		vk = 0xB1
		scan = 0x10
	default:
		return
	}

	user32 := syscall.NewLazyDLL("user32.dll")
	pSendInput := user32.NewProc("SendInput")

	var input kbInput
	input.Type = 1 // INPUT_KEYBOARD
	input.Vk = vk
	input.Scan = scan

	// Key down (Extended key)
	input.Flags = 0x0001 // KEYEVENTF_EXTENDEDKEY
	pSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&input)),
		unsafe.Sizeof(input),
	)

	// Key up (KeyUp + Extended key)
	input.Flags = 0x0002 | 0x0001 // KEYEVENTF_KEYUP | KEYEVENTF_EXTENDEDKEY
	pSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&input)),
		unsafe.Sizeof(input),
	)
}
