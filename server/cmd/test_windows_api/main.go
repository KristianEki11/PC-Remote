//go:build windows

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	winapi "pcremote-server/windows"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println()
		fmt.Println("╔═══════════════════════════════════════════╗")
		fmt.Println("║   PC Remote — Windows API Test Suite      ║")
		fmt.Println("╠═══════════════════════════════════════════╣")
		fmt.Println("║  [1] SetVolume (50%) — Sonar Media        ║")
		fmt.Println("║  [2] GetVolumeStatus — Sonar Media        ║")
		fmt.Println("║  [3] SetMute(true)                        ║")
		fmt.Println("║  [4] SetMute(false)                       ║")
		fmt.Println("║  [5] GetAllChannelVolumes (Sonar)         ║")
		fmt.Println("║  [6] SetChannelVolume (pick channel)      ║")
		fmt.Println("║  [7] SendMediaKey play_pause              ║")
		fmt.Println("║  [8] SendMediaKey next                    ║")
		fmt.Println("║  [9] SendMediaKey prev                    ║")
		fmt.Println("║  [L] LockWorkstation (CAUTION!)           ║")
		fmt.Println("║  [B] OpenBrowser (https://example.com)    ║")
		fmt.Println("║  [0] Exit                                 ║")
		fmt.Println("╚═══════════════════════════════════════════╝")
		fmt.Print("Select option: ")

		if !scanner.Scan() {
			break
		}
		choice := strings.TrimSpace(strings.ToLower(scanner.Text()))

		switch choice {
		case "1":
			fmt.Print("Testing SetVolume(0.5)... ")
			if err := winapi.SetVolume(0.5); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "2":
			fmt.Print("Testing GetVolumeStatus... ")
			level, muted, err := winapi.GetVolumeStatus()
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
			} else {
				fmt.Printf("OK — level=%.2f, muted=%v\n", level, muted)
			}

		case "3":
			fmt.Print("Testing SetMute(true)... ")
			if err := winapi.SetMute(true); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "4":
			fmt.Print("Testing SetMute(false)... ")
			if err := winapi.SetMute(false); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "5":
			fmt.Print("Testing GetAllChannelVolumes... ")
			channels, err := winapi.GetAllChannelVolumes()
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
			} else {
				fmt.Println("OK")
				for k, v := range channels {
					fmt.Printf("  %-8s → level=%.2f  muted=%v  (%s)\n", k, v.Level, v.Muted, v.Name)
				}
				if len(channels) == 0 {
					fmt.Println("  (no Sonar channels detected)")
				}
			}

		case "d":
			fmt.Println("Dumping all discovered audio devices...")
			devs, err := winapi.DebugListAllDevices()
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
			} else {
				for _, d := range devs {
					fmt.Printf("  Device: %q\n", d)
				}
			}

		case "6":
			fmt.Print("Channel name (gaming/chat/media/mic): ")
			if !scanner.Scan() {
				break
			}
			ch := strings.TrimSpace(scanner.Text())
			fmt.Printf("Testing SetChannelVolume(%s, 0.5)... ", ch)
			if err := winapi.SetChannelVolume(ch, 0.5); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "7":
			fmt.Print("Testing SendMediaKey(play_pause)... ")
			if err := winapi.SendMediaKey("play_pause"); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "8":
			fmt.Print("Testing SendMediaKey(next)... ")
			if err := winapi.SendMediaKey("next"); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "9":
			fmt.Print("Testing SendMediaKey(prev)... ")
			if err := winapi.SendMediaKey("prev"); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "l":
			fmt.Println("⚠  WARNING: This will LOCK your screen!")
			fmt.Print("Type 'yes' to confirm: ")
			if scanner.Scan() && strings.TrimSpace(scanner.Text()) == "yes" {
				if err := winapi.LockWorkstation(); err != nil {
					fmt.Printf("ERROR: %v\n", err)
				} else {
					fmt.Println("OK")
				}
			} else {
				fmt.Println("Skipped.")
			}

		case "b":
			fmt.Print("Testing OpenBrowser(https://example.com)... ")
			if err := winapi.OpenBrowser("https://example.com"); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "0":
			fmt.Println("Bye!")
			return

		default:
			fmt.Println("Invalid option, try again.")
		}
	}
}
