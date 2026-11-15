package webrtc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Afrouper/steinel-cam-bridge/pkg/events"
	"github.com/Afrouper/steinel-cam-bridge/pkg/logger"
)

// Timestamp overlay positions of the L 625 (set_time_setting, action "OSD").
// Verified on firmware 0.1.2.31 by looking at the picture.
const (
	OSDOff         = 0
	OSDTopLeft     = 1
	OSDBottomLeft  = 2
	OSDTopRight    = 3
	OSDBottomRight = 4
)

const cameraInfoInterval = 60 * time.Second

// runCameraInfoLoop asks the camera SoC for its device info and its detection
// and recording settings right after the session is up and then once a
// minute. The replies carry the Wi-Fi signal, SD card usage, firmware version
// and update offer, and the camera-side settings.
func (b *Bridge) runCameraInfoLoop(ctx context.Context) {
	ticker := time.NewTicker(cameraInfoInterval)
	defer ticker.Stop()

	b.requestCameraInfo()
	b.requestDetectionSettings()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.requestCameraInfo()
			b.requestDetectionSettings()
		}
	}
}

func (b *Bridge) requestCameraInfo() {
	if err := b.sendJSONCmd("get_device_info", nil); err != nil {
		logger.Debug("DataChannel", "get_device_info could not be sent: %v", err)
	}
}

// refreshCameraInfoSoon re-reads the device info shortly after a setting was
// changed, so the reported state reflects what the camera really stored.
func (b *Bridge) refreshCameraInfoSoon() {
	afterSettingDelay(b.requestCameraInfo)
}

// afterSettingDelay runs fn once the camera has had time to store a change.
func afterSettingDelay(fn func()) {
	time.AfterFunc(1500*time.Millisecond, fn)
}

// onDeviceInfo stores the content of a get_device_info reply in the device status.
// The reply also names the Wi-Fi network; it must never be logged.
func (b *Bridge) onDeviceInfo(info map[string]interface{}) {
	status := b.eventBus.GetStatus()
	applyDeviceInfo(&status, info)
	b.eventBus.UpdateStatus(status)
}

// applyDeviceInfo copies the known fields of a get_device_info reply into st.
// Missing fields leave the previous values untouched.
func applyDeviceInfo(st *events.DeviceStatus, info map[string]interface{}) {
	if fw, ok := info["FW_version"].(string); ok && fw != "" {
		st.FirmwareVer = fw
	}
	if wifi, ok := info["wifi"].(map[string]interface{}); ok {
		if rssi, ok := jsonInt(wifi["RSSI"]); ok {
			st.WifiSignal = rssi
		}
	}
	if sd, ok := info["SD"].(map[string]interface{}); ok {
		if free, ok := jsonInt(sd["volume_free"]); ok {
			st.SDFreeMB = free
		}
		if total, ok := jsonInt(sd["volume_total"]); ok {
			st.SDTotalMB = total
		}
		if s, ok := sd["status"].(string); ok {
			st.SDStatus = s
		}
	}
	if t, ok := info["time"].(map[string]interface{}); ok {
		if osd, ok := jsonInt(t["OSD"]); ok {
			st.OSDPosition = osd
		}
	}
	if vol, ok := jsonInt(info["volume"]); ok {
		st.SpeakerVolume = vol
	}
	if det, ok := info["detect"].(map[string]interface{}); ok {
		if mode, ok := det["mode"].(string); ok {
			st.DetectMode = mode
		}
	}
	applyUpdateInfo(st, info)
}

// jsonInt reads a JSON number (decoded as float64) as int.
func jsonInt(v interface{}) (int, bool) {
	f, ok := v.(float64)
	if !ok {
		return 0, false
	}
	return int(f), true
}

// SetTimestampPosition switches the timestamp overlay off (0) or moves it to
// one of the four corners (see the OSD constants).
func (b *Bridge) SetTimestampPosition(position int) error {
	if position < OSDOff || position > OSDBottomRight {
		return fmt.Errorf("invalid timestamp position %d (0 = off, 1-4 = corners)", position)
	}
	if err := b.sendJSONCmd("set_time_setting", map[string]interface{}{"action": "OSD", "OSD": position}); err != nil {
		return err
	}
	status := b.eventBus.GetStatus()
	status.OSDPosition = position
	b.eventBus.UpdateStatus(status)
	b.refreshCameraInfoSoon()
	return nil
}

// SetSpeakerVolume sets the volume of the camera's speaker (0-100).
func (b *Bridge) SetSpeakerVolume(percent int) error {
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}
	if err := b.sendJSONCmd("set_volume_setting", map[string]interface{}{"volume": percent}); err != nil {
		return err
	}
	status := b.eventBus.GetStatus()
	status.SpeakerVolume = percent
	b.eventBus.UpdateStatus(status)
	b.refreshCameraInfoSoon()
	return nil
}

// SetDetectMode sets the camera-side detection mode (e.g. "ai").
// The firmware stores any text without checking it, so only known values are passed on.
func (b *Bridge) SetDetectMode(mode string) error {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if !IsKnownDetectMode(mode) {
		return fmt.Errorf("unknown detection mode %q", mode)
	}
	if err := b.sendJSONCmd("set_detect_setting", map[string]interface{}{"action": "mode", "mode": mode}); err != nil {
		return err
	}
	status := b.eventBus.GetStatus()
	status.DetectMode = mode
	b.eventBus.UpdateStatus(status)
	b.refreshCameraInfoSoon()
	return nil
}

// DetectModes lists the detection modes the bridge passes to the camera.
var DetectModes = []string{"ai", "off"}

// IsKnownDetectMode reports whether mode is one of DetectModes.
func IsKnownDetectMode(mode string) bool {
	for _, m := range DetectModes {
		if m == mode {
			return true
		}
	}
	return false
}
