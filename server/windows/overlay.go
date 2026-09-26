package windows

import (
	"log/slog"
	"runtime"
	"syscall"
	"unsafe"
	"github.com/lxn/win"
)

var (
    overlayHwnd win.HWND
    isOverlayRunning bool
    transparentCursor win.HCURSOR
)

// ShowBlackOverlay opens a full-screen black window to simulate display off.
func ShowBlackOverlay() {
    if isOverlayRunning {
        return
    }
    isOverlayRunning = true
    go showBlackOverlayThread()
}

// showBlackOverlayThread must run on its own locked thread because it runs a message loop.
func showBlackOverlayThread() {
    runtime.LockOSThread()
    defer runtime.UnlockOSThread()
    
	hInstance := win.GetModuleHandle(nil)
	className := syscall.StringToUTF16Ptr("PCRemoteBlackOverlayClass")

	// Create invisible transparent cursor
	andMask := make([]byte, 128)
	for i := range andMask {
		andMask[i] = 0xFF
	}
	xorMask := make([]byte, 128)
	user32 := syscall.NewLazyDLL("user32.dll")
	procCreateCursor := user32.NewProc("CreateCursor")
	retCursor, _, _ := procCreateCursor.Call(
		uintptr(hInstance),
		0, 0,
		32, 32,
		uintptr(unsafe.Pointer(&andMask[0])),
		uintptr(unsafe.Pointer(&xorMask[0])),
	)
	transparentCursor = win.HCURSOR(retCursor)

	wc := win.WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(win.WNDCLASSEX{})),
		HInstance:     hInstance,
		LpszClassName: className,
		LpfnWndProc:   syscall.NewCallback(overlayWndProc),
		HbrBackground: win.HBRUSH(win.GetStockObject(win.BLACK_BRUSH)),
		HCursor:       transparentCursor,
	}

	win.RegisterClassEx(&wc)

	x := win.GetSystemMetrics(win.SM_XVIRTUALSCREEN)
	y := win.GetSystemMetrics(win.SM_YVIRTUALSCREEN)
	w := win.GetSystemMetrics(win.SM_CXVIRTUALSCREEN)
	h := win.GetSystemMetrics(win.SM_CYVIRTUALSCREEN)

	overlayHwnd = win.CreateWindowEx(
		win.WS_EX_TOPMOST|win.WS_EX_TOOLWINDOW,
		className,
		syscall.StringToUTF16Ptr("BlackOverlay"),
		win.WS_POPUP|win.WS_VISIBLE,
		x, y, w, h,
		0, 0, hInstance, nil,
	)

	if overlayHwnd == 0 {
		slog.Error("Failed to create overlay window")
        isOverlayRunning = false
		return
	}

	win.SetForegroundWindow(overlayHwnd)
	win.SetFocus(overlayHwnd)
	win.SetCapture(overlayHwnd)

	// Fallback: move the physical cursor off-screen to the bottom right
	win.SetCursorPos(w, h)

	slog.Info("Native Go Overlay window created")

    

	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) != 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}

    

    isOverlayRunning = false
    overlayHwnd = 0
}

func closeOverlay() {
	if overlayHwnd != 0 {
		win.PostMessage(overlayHwnd, win.WM_CLOSE, 0, 0)
	}
}

func overlayWndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_SETCURSOR:
		win.SetCursor(transparentCursor)
		return 1
	case win.WM_KEYDOWN, win.WM_LBUTTONDOWN, win.WM_RBUTTONDOWN, win.WM_MBUTTONDOWN:
		win.ReleaseCapture()
		win.PostMessage(hwnd, win.WM_CLOSE, 0, 0)
		return 0
	case win.WM_DESTROY:
		win.PostQuitMessage(0)
		return 0
	}
	return win.DefWindowProc(hwnd, msg, wParam, lParam)
}



