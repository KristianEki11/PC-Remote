package windows

import (
	"log/slog"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

var (
	overlayHwnd      win.HWND
	isOverlayRunning bool
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
	user32                   = syscall.NewLazyDLL("user32.dll")
	procCreateCursor         = user32.NewProc("CreateCursor")
	procSetSystemCursor      = user32.NewProc("SetSystemCursor")
	procSystemParametersInfo = user32.NewProc("SystemParametersInfoW")
	procCopyCursor           = user32.NewProc("CopyIcon") // CopyCursor is a C macro for CopyIcon
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
// This works system-wide regardless of which window has focus.
func hideSystemCursors() {
	blankCursor := createBlankCursor()
	if blankCursor == 0 {
		slog.Error("Failed to create blank cursor")
		return
	}

	for _, id := range allCursorIDs {
		// SetSystemCursor destroys the cursor handle passed to it,
		// so we must create a fresh copy for each call.
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

// ShowBlackOverlay opens a full-screen black window to simulate display off.
func ShowBlackOverlay() {
	if isOverlayRunning {
		return
	}
	isOverlayRunning = true
	go showBlackOverlayThread()
}

// showBlackOverlayThread must run on its own OS-locked thread for the message loop.
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

	// Hide ALL system cursors (system-wide, works even without focus)
	hideSystemCursors()

	slog.Info("Native overlay window created, cursors hidden")

	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) != 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}

	// Restore cursors when overlay closes
	restoreSystemCursors()

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
	case win.WM_DESTROY:
		win.PostQuitMessage(0)
		return 0
	}
	return win.DefWindowProc(hwnd, msg, wParam, lParam)
}
