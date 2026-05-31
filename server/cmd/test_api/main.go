package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Global config
var (
	serverURL string
	pin       string
	reader    *bufio.Reader
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

func main() {
	reader = bufio.NewReader(os.Stdin)

	// 1. Auto-discover PIN and Port
	defaultPin := discoverPin()
	defaultPort := discoverPort()
	serverURL = fmt.Sprintf("http://localhost:%s", defaultPort)
	pin = defaultPin

	// Print startup header
	fmt.Printf("%s============================================================%s\n", colorBlue, colorReset)
	fmt.Printf("%s   PCRemote Server — Interactive API Console Test Tool%s\n", colorBold+colorCyan, colorReset)
	fmt.Printf("%s============================================================%s\n", colorBlue, colorReset)

	fmt.Printf("Masukkan IP Server [default: localhost]: ")
	ipInput, _ := reader.ReadString('\n')
	ipInput = strings.TrimSpace(ipInput)
	if ipInput != "" {
		if !strings.HasPrefix(ipInput, "http://") && !strings.HasPrefix(ipInput, "https://") {
			serverURL = "http://" + ipInput
		} else {
			serverURL = ipInput
		}
		if !strings.Contains(serverURL, ":") {
			serverURL = serverURL + ":" + defaultPort
		}
	}

	fmt.Printf("Masukkan PIN Otentikasi [default: %s]: ", pin)
	pinInput, _ := reader.ReadString('\n')
	pinInput = strings.TrimSpace(pinInput)
	if pinInput != "" {
		pin = pinInput
	}

	for {
		printMenu()
		fmt.Print("\nPilih opsi (0-15): ")
		choiceStr, _ := reader.ReadString('\n')
		choiceStr = strings.TrimSpace(choiceStr)
		if choiceStr == "" {
			continue
		}
		choice, err := strconv.Atoi(choiceStr)
		if err != nil {
			fmt.Printf("%sPilihan tidak valid.%s\n", colorRed, colorReset)
			time.Sleep(1 * time.Second)
			continue
		}

		if choice == 0 {
			fmt.Println("Keluar dari program. Sampai jumpa!")
			break
		}

		handleChoice(choice)
		fmt.Println("\nTekan ENTER untuk kembali ke menu utama...")
		reader.ReadString('\n')
	}
}

func printMenu() {
	fmt.Print("\033[H\033[2J")
	fmt.Printf("%s============================================================%s\n", colorBlue, colorReset)
	fmt.Printf("%s        PC REMOTE DASHBOARD & API TEST CONSOLE%s\n", colorBold+colorCyan, colorReset)
	fmt.Printf("   Target Server: %s%s%s | PIN: %s%s%s\n", colorBold, serverURL, colorReset, colorBold, pin, colorReset)
	fmt.Printf("%s============================================================%s\n", colorBlue, colorReset)
	fmt.Println(" [MANAJEMEN SERVER]")
	fmt.Printf("  %s111. Jalankan PC Remote Server (Background)%s\n", colorGreen, colorReset)
	fmt.Printf("  %s222. Hentikan PC Remote Server%s\n", colorRed, colorReset)
	fmt.Printf("%s------------------------------------------------------------%s\n", colorBlue, colorReset)
	fmt.Println(" [PENGUJIAN API]")
	fmt.Println("  1. Health Check (Tanpa Auth)")
	fmt.Println("  2. Cek Status Audio Utama (Volume, Mute, Device)")
	fmt.Println("  3. Set Volume Utama (0.0 - 1.0)")
	fmt.Println("  4. Toggle Mute Audio Utama")
	fmt.Println("  5. Media: Play / Pause")
	fmt.Println("  6. Media: Next Track")
	fmt.Println("  7. Media: Previous Track")
	fmt.Println("  8. Buka URL di Browser Default PC")
	fmt.Println("  9. Kunci Layar PC (Lock Workstation)")
	fmt.Println(" 10. Masuk Mode Standby/Sleep PC")
	fmt.Println(" 11. Jadwalkan Shutdown PC")
	fmt.Println(" 12. Batalkan Shutdown PC")
	fmt.Println(" 13. Cek Status Saluran Audio (SteelSeries Sonar)")
	fmt.Println(" 14. Set Volume Saluran Audio Sonar")
	fmt.Println(" 15. Toggle Mute Saluran Audio Sonar")
	fmt.Println("  0. Keluar dari Dashboard")
	fmt.Printf("%s============================================================%s\n", colorBlue, colorReset)
}

func handleChoice(choice int) {
	switch choice {
	case 111:
		fmt.Printf("%sMenjalankan Server PC Remote...%s\n", colorYellow, colorReset)
		// Jalankan exe tanpa menahan console (detach)
		exePath := ".\\pcremote-server.exe"
		if _, err := os.Stat(exePath); os.IsNotExist(err) {
			exePath = "C:\\Program Files\\PCRemote\\pcremote-server.exe"
		}
		cmd := exec.Command(exePath)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if err := cmd.Start(); err != nil {
			fmt.Printf("%sGagal menjalankan server: %v%s\n", colorRed, err, colorReset)
		} else {
			fmt.Printf("%s[SUKSES] Server berjalan di background (PID: %d)%s\n", colorGreen, cmd.Process.Pid, colorReset)
		}
	case 222:
		fmt.Printf("%sMematikan Server PC Remote...%s\n", colorYellow, colorReset)
		cmd := exec.Command("taskkill", "/F", "/IM", "pcremote-server.exe")
		if err := cmd.Run(); err != nil {
			fmt.Printf("%sGagal menghentikan server (mungkin tidak berjalan).%s\n", colorYellow, colorReset)
		} else {
			fmt.Printf("%s[SUKSES] Server berhasil dihentikan.%s\n", colorGreen, colorReset)
		}
	case 1:
		sendRequest("GET", "/health", false, nil)
	case 2:
		sendRequest("GET", "/audio/status", true, nil)
	case 3:
		fmt.Print("Masukkan level volume (0.0 - 1.0): ")
		volStr, _ := reader.ReadString('\n')
		vol, err := strconv.ParseFloat(strings.TrimSpace(volStr), 64)
		if err != nil || vol < 0.0 || vol > 1.0 {
			fmt.Printf("%sVolume harus berupa angka desimal antara 0.0 dan 1.0!%s\n", colorRed, colorReset)
			return
		}
		sendRequest("POST", "/audio/volume", true, map[string]any{"level": vol})
	case 4:
		fmt.Print("Mute? (y/n): ")
		muteStr, _ := reader.ReadString('\n')
		muted := strings.ToLower(strings.TrimSpace(muteStr)) == "y"
		sendRequest("POST", "/audio/mute", true, map[string]any{"muted": muted})
	case 5:
		sendRequest("POST", "/media/play", true, nil)
	case 6:
		sendRequest("POST", "/media/next", true, nil)
	case 7:
		sendRequest("POST", "/media/prev", true, nil)
	case 8:
		fmt.Print("Masukkan URL (contoh: https://youtube.com): ")
		urlStr, _ := reader.ReadString('\n')
		urlStr = strings.TrimSpace(urlStr)
		if urlStr == "" {
			fmt.Printf("%sURL tidak boleh kosong!%s\n", colorRed, colorReset)
			return
		}
		sendRequest("POST", "/browser/open", true, map[string]any{"url": urlStr})
	case 9:
		sendRequest("POST", "/system/lock", true, nil)
	case 10:
		fmt.Print("Apakah Anda yakin ingin menidurkan PC? (y/n): ")
		confirm, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
			fmt.Println("Dibatalkan.")
			return
		}
		sendRequest("POST", "/system/sleep", true, nil)
	case 11:
		fmt.Print("Masukkan waktu tunda shutdown (dalam menit): ")
		delayStr, _ := reader.ReadString('\n')
		delay, err := strconv.Atoi(strings.TrimSpace(delayStr))
		if err != nil || delay < 0 {
			fmt.Printf("%sMenit harus berupa angka bulat positif!%s\n", colorRed, colorReset)
			return
		}
		sendRequest("POST", "/system/shutdown", true, map[string]any{"delay_seconds": delay * 60})
	case 12:
		sendRequest("POST", "/system/shutdown/cancel", true, nil)
	case 13:
		sendRequest("GET", "/audio/devices", true, nil)
	case 14:
		fmt.Print("Masukkan ID/Nama Saluran Sonar (gaming, chat, media, mic): ")
		ch, _ := reader.ReadString('\n')
		ch = strings.TrimSpace(ch)
		fmt.Print("Masukkan level volume (0.0 - 1.0): ")
		volStr, _ := reader.ReadString('\n')
		vol, err := strconv.ParseFloat(strings.TrimSpace(volStr), 64)
		if err != nil || vol < 0.0 || vol > 1.0 {
			fmt.Printf("%sVolume harus berupa angka desimal antara 0.0 dan 1.0!%s\n", colorRed, colorReset)
			return
		}
		sendRequest("POST", "/audio/device/volume", true, map[string]any{"device_id": ch, "level": vol})
	case 15:
		fmt.Print("Masukkan ID/Nama Saluran Sonar (gaming, chat, media, mic): ")
		ch, _ := reader.ReadString('\n')
		ch = strings.TrimSpace(ch)
		fmt.Print("Mute? (y/n): ")
		muteStr, _ := reader.ReadString('\n')
		muted := strings.ToLower(strings.TrimSpace(muteStr)) == "y"
		sendRequest("POST", "/audio/device/mute", true, map[string]any{"device_id": ch, "mute": muted})
	default:
		fmt.Printf("%sPilihan menu tidak tersedia.%s\n", colorRed, colorReset)
	}
}

func sendRequest(method string, endpoint string, usePin bool, payload any) {
	fmt.Printf("\n%sMengirim request %s ke %s...%s\n", colorYellow, method, endpoint, colorReset)

	var body io.Reader
	if payload != nil {
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			fmt.Printf("%s[ERR] JSON marshal failed: %v%s\n", colorRed, err, colorReset)
			return
		}
		body = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, serverURL+endpoint, body)
	if err != nil {
		fmt.Printf("%s[ERR] Gagal membuat request: %v%s\n", colorRed, err, colorReset)
		return
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if usePin {
		req.Header.Set("X-PIN", pin)
	}

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s[ERR] Koneksi ke server gagal: %v%s\n", colorRed, err, colorReset)
		return
	}
	defer resp.Body.Close()

	respBodyBytes, _ := io.ReadAll(resp.Body)
	respBody := string(respBodyBytes)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("%s[SUKSES] HTTP %d (%s)%s\n", colorGreen, resp.StatusCode, resp.Status, colorReset)
	} else {
		fmt.Printf("%s[GAGAL] HTTP %d (%s)%s\n", colorRed, resp.StatusCode, resp.Status, colorReset)
	}

	// Pretty print JSON response if possible
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, respBodyBytes, "", "  "); err == nil {
		fmt.Printf("Response Body:\n%s\n", prettyJSON.String())
	} else {
		if strings.TrimSpace(respBody) != "" {
			fmt.Printf("Response Body: %s\n", respBody)
		} else {
			fmt.Println("Response Body: (kosong)")
		}
	}
}

func discoverPin() string {
	paths := []string{
		".env",
		"server/.env",
		`C:\Program Files\PCRemote\.env`,
	}
	for _, p := range paths {
		if content, err := os.ReadFile(p); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "PIN=") {
					return strings.TrimPrefix(line, "PIN=")
				}
				if strings.HasPrefix(line, "APP_PIN=") {
					return strings.TrimPrefix(line, "APP_PIN=")
				}
			}
		}
	}
	return "1234"
}

func discoverPort() string {
	if p := os.Getenv("APP_PORT"); p != "" {
		return p
	}
	paths := []string{
		".env",
		"server/.env",
		`C:\Program Files\PCRemote\.env`,
	}
	for _, p := range paths {
		if content, err := os.ReadFile(p); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "PORT=") {
					return strings.TrimPrefix(line, "PORT=")
				}
				if strings.HasPrefix(line, "APP_PORT=") {
					return strings.TrimPrefix(line, "APP_PORT=")
				}
			}
		}
	}
	return "8000"
}
