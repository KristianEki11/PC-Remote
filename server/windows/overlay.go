package windows

import (
	"log/slog"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

var (
	overlayHwnd      win.HWND
	isOverlayRunning bool
	overlayMutex     sync.Mutex
	currentAlpha     byte = 255
	isClickThrough   bool = false
)

// Win32 cursor type IDs for SetSystemCursor
const (
	ocrNormal      = 32512
	ocrIbeam       = 32513
	ocrWait        = 32514
	ocrCross       = 32515
	ocrUp          = 32516
	ocrSizeNWSE    = 32642
	ocrSizeNESW    = 32643
	ocrSizeWE      = 32644
	ocrSizeNS      = 32645
	ocrSizeAll     = 32646
	ocrNo          = 32648
	ocrHand        = 32649
	ocrAppStarting = 32650
	spiSetCursors  = 0x0057
)

var allCursorIDs = []uintptr{
	ocrNormal, ocrIbeam, ocrWait, ocrCross, ocrUp,
	ocrSizeNWSE, ocrSizeNESW, ocrSizeWE, ocrSizeNS, ocrSizeAll,
	ocrNo, ocrHand, ocrAppStarting,
}

var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	procCreateCursor             = user32.NewProc("CreateCursor")
	procSetSystemCursor          = user32.NewProc("SetSystemCursor")
	procSystemParametersInfo     = user32.NewProc("SystemParametersInfoW")
	procCopyCursor               = user32.NewProc("CopyIcon") // CopyCursor is a C macro for CopyIcon
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	procSetWindowLong            = user32.NewProc("SetWindowLongW")
	procGetWindowLong            = user32.NewProc("GetWindowLongW")
)

const (
	gwlExStyle        = uint32(0xFFFFFFEC) // -20 in two's complement
	wsExLayered       = 0x00080000
	wsExTransparent   = 0x00000020
	lwaAlpha          = 0x00000002
	wmUserSetAlpha    = win.WM_USER + 101
	wmUserClose       = win.WM_USER + 102
)

// createBlankCursor creates a fully transparent 32x32 cursor in memory.
func createBlankCursor() win.HCURSOR {
	andMask := make([]byte, 128) // 32x32 / 8 = 128 bytes
	for i := range andMask {
		andMask[i] = 0xFF // AND mask: all 1s = transparent
	}
	xorMask := make([]byte, 128) // XOR mask: all 0s = no inversion

	ret, _, _ := procCreateCursor.Call(
		0, // hInstance (0 is fine for in-memory cursors)
		0, 0, // hotspot x, y
		32, 32, // width, height
		uintptr(unsafe.Pointer(&andMask[0])),
		uintptr(unsafe.Pointer(&xorMask[0])),
	)
	return win.HCURSOR(ret)
}

// hideSystemCursors replaces ALL system cursor types with a blank cursor.
func hideSystemCursors() {
	blankCursor := createBlankCursor()
	if blankCursor == 0 {
		slog.Error("Failed to create blank cursor")
		return
	}

	for _, id := range allCursorIDs {
		copied, _, _ := procCopyCursor.Call(uintptr(blankCursor))
		if copied != 0 {
			procSetSystemCursor.Call(copied, id)
		}
	}
	slog.Info("System cursors hidden (all replaced with blank)")
}

// restoreSystemCursors reloads the default cursor scheme from the registry.
func restoreSystemCursors() {
	procSystemParametersInfo.Call(spiSetCursors, 0, 0, 0)
	slog.Info("System cursors restored")
}

// ShowBlackOverlay opens a full-screen black window to simulate display off (100% black, intercepts input).
func ShowBlackOverlay() {
	SetOverlayOpacity(255, false)
}

// SetOverlayOpacity updates or shows the overlay with a given alpha (0-255).
// clickThrough: true allows mouse/keyboard events to pass straight through the window.
func SetOverlayOpacity(alpha byte, clickThrough bool) {
	overlayMutex.Lock()
	defer overlayMutex.Unlock()

	currentAlpha = alpha
	isClickThrough = clickThrough

	if alpha == 0 {
		if overlayHwnd != 0 {
			win.PostMessage(overlayHwnd, wmUserClose, 0, 0)
		}
		return
	}

	if overlayHwnd != 0 {
		// Update running window
		win.PostMessage(overlayHwnd, wmUserSetAlpha, uintptr(alpha), 0)
		return
	}

	if !isOverlayRunning {
		isOverlayRunning = true
		go showOverlayThread(alpha, clickThrough)
	}
}

// CloseOverlay closes the overlay if active.
func CloseOverlay() {
	overlayMutex.Lock()
	defer overlayMutex.Unlock()

	if overlayHwnd != 0 {
		win.PostMessage(overlayHwnd, wmUserClose, 0, 0)
	}
}

func showOverlayThread(initAlpha byte, clickThrough bool) {
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
	}

	win.RegisterClassEx(&wc)

	x := win.GetSystemMetrics(win.SM_XVIRTUALSCREEN)
	y := win.GetSystemMetrics(win.SM_YVIRTUALSCREEN)
	w := win.GetSystemMetrics(win.SM_CXVIRTUALSCREEN)
	h := win.GetSystemMetrics(win.SM_CYVIRTUALSCREEN)

	exStyle := uint32(win.WS_EX_TOPMOST | win.WS_EX_TOOLWINDOW | wsExLayered)
	if clickThrough {
		exStyle |= wsExTransparent
	}

	hwnd := win.CreateWindowEx(
		exStyle,
		className,
		syscall.StringToUTF16Ptr("BlackOverlay"),
		win.WS_POPUP | win.WS_VISIBLE,
		x, y, w, h,
		0, 0, hInstance, nil,
	)

	if hwnd == 0 {
		slog.Error("Failed to create overlay window")
		overlayMutex.Lock()
		isOverlayRunning = false
		overlayMutex.Unlock()
		return
	}

	overlayMutex.Lock()
	overlayHwnd = hwnd
	overlayMutex.Unlock()

	// Apply initial alpha
	procSetLayeredWindowAttributes.Call(uintptr(hwnd), 0, uintptr(initAlpha), uintptr(lwaAlpha))

	if !clickThrough && initAlpha == 255 {
		hideSystemCursors()
	}

	slog.Info("Native overlay window created", "alpha", initAlpha, "clickThrough", clickThrough)

	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) != 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}

	restoreSystemCursors()

	overlayMutex.Lock()
	isOverlayRunning = false
	overlayHwnd = 0
	overlayMutex.Unlock()
}

func overlayWndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmUserSetAlpha:
		alpha := byte(wParam)
		overlayMutex.Lock()
		clickThrough := isClickThrough
		overlayMutex.Unlock()

		// Update style for click-through if necessary
		curEx, _, _ := procGetWindowLong.Call(uintptr(hwnd), uintptr(gwlExStyle))
		newEx := curEx | wsExLayered
		if clickThrough {
			newEx |= wsExTransparent
		} else {
			newEx &^= wsExTransparent
		}
		if newEx != curEx {
			procSetWindowLong.Call(uintptr(hwnd), uintptr(gwlExStyle), newEx)
		}

		procSetLayeredWindowAttributes.Call(uintptr(hwnd), 0, uintptr(alpha), uintptr(lwaAlpha))
		return 0

	case wmUserClose:
		win.DestroyWindow(hwnd)
		return 0

	case win.WM_KEYDOWN, win.WM_LBUTTONDOWN, win.WM_RBUTTONDOWN, win.WM_MBUTTONDOWN:
		overlayMutex.Lock()
		isFullBlack := !isClickThrough && currentAlpha == 255
		overlayMutex.Unlock()
		if isFullBlack {
			win.DestroyWindow(hwnd)
		}
		return 0

	case win.WM_DESTROY:
		win.PostQuitMessage(0)
		return 0
	}
	return win.DefWindowProc(hwnd, msg, wParam, lParam)
}
