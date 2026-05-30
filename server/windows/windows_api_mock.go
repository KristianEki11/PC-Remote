//go:build !windows

package windows

import (
	"log"
)

type MockAPI struct{}

func (MockAPI) SetVolume(level float64) error {
	log.Printf("MOCK: SetVolume called with %f", level)
	return nil
}

func (MockAPI) GetVolumeStatus() (float64, bool, error) {
	log.Println("MOCK: GetVolumeStatus called")
	return 0.5, false, nil
}

func (MockAPI) SetMute(muted bool) error {
	log.Printf("MOCK: SetMute called with %v", muted)
	return nil
}

func (MockAPI) GetAllChannelVolumes() (map[string]ChannelStatus, error) {
	log.Println("MOCK: GetAllChannelVolumes called")
	return map[string]ChannelStatus{
		"media": {Name: "Sonar - Media", Level: 0.5, Muted: false},
	}, nil
}

func (MockAPI) SetChannelVolume(channelName string, level float64) error {
	log.Printf("MOCK: SetChannelVolume called with %s, %f", channelName, level)
	return nil
}

func (MockAPI) SetChannelMute(channelName string, muted bool) error {
	log.Printf("MOCK: SetChannelMute called with %s, %v", channelName, muted)
	return nil
}

func (MockAPI) GetDevices() ([]DeviceStatus, error) {
	log.Println("MOCK: GetDevices called")
	return []DeviceStatus{
		{ID: "mock-device-id", Name: "Mock Speakers", Volume: 50, Muted: false},
	}, nil
}

func (MockAPI) SetDeviceVolume(id string, level float64) error {
	log.Printf("MOCK: SetDeviceVolume called with %s, %f", id, level)
	return nil
}

func (MockAPI) SetDeviceMute(id string, muted bool) error {
	log.Printf("MOCK: SetDeviceMute called with %s, %v", id, muted)
	return nil
}

func (MockAPI) SendMediaKey(action string) error {
	log.Printf("MOCK: SendMediaKey called with %s", action)
	return nil
}

func (MockAPI) LockWorkstation() error {
	log.Println("MOCK: LockWorkstation called")
	return nil
}

func (MockAPI) OpenBrowser(url string) error {
	log.Printf("MOCK: OpenBrowser called with %s", url)
	return nil
}

func (MockAPI) ScheduleShutdown(delaySeconds int) error {
	log.Printf("MOCK: ScheduleShutdown called with %d seconds", delaySeconds)
	return nil
}

func (MockAPI) CancelShutdown() error {
	log.Println("MOCK: CancelShutdown called")
	return nil
}

func (MockAPI) Sleep() error {
	log.Println("MOCK: Sleep called")
	return nil
}

func (MockAPI) Restart() error {
	log.Println("MOCK: Restart called")
	return nil
}

func (MockAPI) DebugListAllDevices() ([]string, error) {
	log.Println("MOCK: DebugListAllDevices called")
	return []string{"[MOCK] Speakers"}, nil
}

func init() {
	API = MockAPI{}
}

// Package-level functions delegating to API
		func SetVolume(level float64) error { return API.SetVolume(level) }
		func GetVolumeStatus() (float64, bool, error) { return API.GetVolumeStatus() }
		func SetMute(muted bool) error { return API.SetMute(muted) }
		func GetAllChannelVolumes() (map[string]ChannelStatus, error) { return API.GetAllChannelVolumes() }
		func SetChannelVolume(channelName string, level float64) error { return API.SetChannelVolume(channelName, level) }
		func SetChannelMute(channelName string, muted bool) error { return API.SetChannelMute(channelName, muted) }
		func GetDevices() ([]DeviceStatus, error) { return API.GetDevices() }
		func SetDeviceVolume(id string, level float64) error { return API.SetDeviceVolume(id, level) }
		func SetDeviceMute(id string, muted bool) error { return API.SetDeviceMute(id, muted) }
		func SendMediaKey(action string) error { return API.SendMediaKey(action) }
		func LockWorkstation() error { return API.LockWorkstation() }
		func OpenBrowser(url string) error { return API.OpenBrowser(url) }
		func ScheduleShutdown(delaySeconds int) error { return API.ScheduleShutdown(delaySeconds) }
		func CancelShutdown() error { return API.CancelShutdown() }
		func Sleep() error { return API.Sleep() }
		func Restart() error { return API.Restart() }
		func DebugListAllDevices() ([]string, error) { return API.DebugListAllDevices() }
