package windows

type ChannelStatus struct {
	Name  string  `json:"name"`
	Level float64 `json:"level"`
	Muted bool    `json:"muted"`
}

type DeviceStatus struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Volume float64 `json:"volume"` // 0-100 to match flutter
	Muted  bool    `json:"muted"`
}

type APIInterface interface {
	SetVolume(level float64) error
	GetVolumeStatus() (level float64, muted bool, err error)
	SetMute(muted bool) error
	GetAllChannelVolumes() (map[string]ChannelStatus, error)
	SetChannelVolume(channelName string, level float64) error
	SetChannelMute(channelName string, muted bool) error
	GetDevices() ([]DeviceStatus, error)
	SetDeviceVolume(id string, level float64) error
	SetDeviceMute(id string, muted bool) error
	SendMediaKey(action string) error
	GetMediaStatus() (map[string]any, error)
	LockWorkstation() error
	OpenBrowser(url string) error
	ScheduleShutdown(delaySeconds int) error
	CancelShutdown() error
	Sleep() error
	Restart() error
	TurnOffDisplay() error
	DebugListAllDevices() ([]string, error)
}

var API APIInterface
