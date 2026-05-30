import multiprocessing
multiprocessing.freeze_support()

import sys
import ctypes

# ──────────────────────────────────────
# Windows Single Instance Mutex Protection
# ──────────────────────────────────────
MUTEX_NAME = "Global\\PCRemoteServer_SingleInstance_Mutex"
# We store the mutex handle globally so it is not garbage collected
single_instance_mutex = ctypes.windll.kernel32.CreateMutexW(None, True, MUTEX_NAME)
last_error = ctypes.windll.kernel32.GetLastError()
if last_error == 183:  # ERROR_ALREADY_EXISTS
    # Show native Windows MessageBox informing the user and exit immediately
    ctypes.windll.user32.MessageBoxW(
        0, 
        "PC Remote Server is already running. Please check your system tray.", 
        "PC Remote Server", 
        0x00000040 | 0x00000000  # MB_OK | MB_ICONINFORMATION
    )
    sys.exit(0)

# ──────────────────────────────────────
# Detach/Hide Console Window (For EXE runs)
# ──────────────────────────────────────
if getattr(sys, 'frozen', False):
    hwnd = ctypes.windll.kernel32.GetConsoleWindow()
    if hwnd:
        ctypes.windll.user32.ShowWindow(hwnd, 0)
    ctypes.windll.kernel32.FreeConsole()

import threading
import time
import webbrowser
import os
import uvicorn
from pystray import Icon, Menu, MenuItem
from PIL import Image, ImageDraw

# Set working directory to directory containing this .exe/script
if getattr(sys, 'frozen', False):
    BASE_DIR = os.path.dirname(sys.executable)
else:
    BASE_DIR = os.path.dirname(os.path.abspath(__file__))

os.chdir(BASE_DIR)

# Import the FastAPI application
try:
    from main import app
except Exception as e:
    # If import fails, we cannot proceed, log to dialog box
    ctypes.windll.user32.MessageBoxW(0, f"Error importing application: {e}", "PC Remote Server Startup Error", 0x10)
    sys.exit(1)

class ServerManager:
    def __init__(self):
        self.thread = None
        self.server = None

    def start(self):
        try:
            if self.is_running():
                return
            
            config = uvicorn.Config(
                app,
                host="0.0.0.0",
                port=8000,
                loop="asyncio",
                workers=1,
                log_config=None,
                use_colors=False
            )
            self.server = uvicorn.Server(config)
            
            self.thread = threading.Thread(target=self.server.run, name="PCRemoteUvicornServer", daemon=True)
            self.thread.start()
        except Exception as e:
            ctypes.windll.user32.MessageBoxW(0, f"Error starting server: {e}", "Server Error", 0x10)

    def stop(self):
        try:
            if self.server:
                self.server.should_exit = True
            if self.thread:
                self.thread.join(timeout=3.0)
        except Exception as e:
            pass

    def is_running(self):
        return self.thread is not None and self.thread.is_alive()

server_manager = ServerManager()

def create_image():
    """Load favicon or create a simple fallback square icon."""
    try:
        icon_path = os.path.join(BASE_DIR, "favicon.ico")
        if os.path.exists(icon_path):
            return Image.open(icon_path)
            
        # Fallback icon: Blue square 64x64
        image = Image.new('RGB', (64, 64), color=(0, 120, 215))
        return image
    except Exception:
        # Fallback image
        return Image.new('RGB', (64, 64), color=(0, 120, 215))

def get_menu():
    """Build the system tray menu dynamically based on server status."""
    status_text = "Status: Running" if server_manager.is_running() else "Status: Stopped"
    return Menu(
        MenuItem("PC Remote Server", None, enabled=False),
        MenuItem(Menu.SEPARATOR, None),
        MenuItem(status_text, None, enabled=False),
        MenuItem("Buka Docs", lambda icon, item: open_docs()),
        MenuItem(Menu.SEPARATOR, None),
        MenuItem("Restart Server", lambda icon, item: restart_server(icon)),
        MenuItem("Exit", lambda icon, item: exit_app(icon))
    )

def open_docs():
    """Open FastAPI Swagger docs in default browser."""
    try:
        webbrowser.open("http://localhost:8000/docs")
    except Exception:
        pass

def restart_server(icon):
    """Restart FastAPI Uvicorn server."""
    try:
        server_manager.stop()
        time.sleep(1.0)
        server_manager.start()
        update_menu(icon)
    except Exception:
        pass

def exit_app(icon):
    """Shutdown server and stop tray loop."""
    try:
        server_manager.stop()
        icon.stop()
    except Exception:
        pass
    finally:
        sys.exit(0)

def update_menu(icon):
    icon.menu = get_menu()

def status_monitor(icon):
    """Monitor thread to update menu text dynamically if server state changes."""
    try:
        last_status = server_manager.is_running()
        while True:
            time.sleep(2.0)
            current_status = server_manager.is_running()
            if current_status != last_status:
                last_status = current_status
                update_menu(icon)
    except Exception:
        pass

def main():
    try:
        # 1. Start uvicorn server thread
        server_manager.start()

        # 2. Initialize tray icon
        icon_image = create_image()
        icon = Icon("PCRemote", icon_image, title="PC Remote Server")
        icon.menu = get_menu()

        # 3. Start state monitor thread
        monitor_thread = threading.Thread(target=status_monitor, args=(icon,), name="PCRemoteStatusMonitor", daemon=True)
        monitor_thread.start()

        # 4. Start blocking GUI event loop
        icon.run()
    except Exception as e:
        ctypes.windll.user32.MessageBoxW(0, f"Tray application crash: {e}", "Fatal Error", 0x10)
        sys.exit(1)

if __name__ == "__main__":
    main()
