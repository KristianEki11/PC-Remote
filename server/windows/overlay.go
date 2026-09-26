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

	wc := win.WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(win.WNDCLASSEX{})),
		HInstance:     hInstance,
		LpszClassName: className,
		LpfnWndProc:   syscall.NewCallback(overlayWndProc),
		HbrBackground: win.HBRUSH(win.GetStockObject(win.BLACK_BRUSH)),
		HCursor:       0,
	}

	win.RegisterClassEx(&wc)

	x := win.GetSystemMetrics(win.SM_XVIRTUALSCREEN)
	y := win.GetSystemMetrics(win.SM_YVIRTUALSCREEN)
	w := win.GetSystemMetrics(win.SM_CXVIRTUALSCREEN)
	h := win.GetSystemMetrics(win.SM_CYVIRTUALSCREEN)

	overlayHwnd = win.CreateWindowEx(
		win.WS_EX_TOPMOST|win.WS_EX_TOOLWINDOW|win.WS_EX_NOACTIVATE,
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
	slog.Info("Native Go Overlay window created")

    // Hide the cursor for this thread's windows
    user32 := syscall.NewLazyDLL("user32.dll")
    procShowCursor := user32.NewProc("ShowCursor")
    for {
        ret, _, _ := procShowCursor.Call(0)
        if int32(ret) < 0 {
            break
        }
    }

	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) != 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}

    // Restore cursor
    for {
        ret, _, _ := procShowCursor.Call(1)
        if int32(ret) >= 0 {
            break
        }
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
	case win.WM_KEYDOWN, win.WM_LBUTTONDOWN, win.WM_RBUTTONDOWN, win.WM_MBUTTONDOWN:
		win.PostMessage(hwnd, win.WM_CLOSE, 0, 0)
		return 0
	case win.WM_SETCURSOR:
		win.SetCursor(0)
		return 1
	case win.WM_DESTROY:
		win.PostQuitMessage(0)
		return 0
	}
	return win.DefWindowProc(hwnd, msg, wParam, lParam)
}
