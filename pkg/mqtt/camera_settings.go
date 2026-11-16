package mqtt

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Afrouper/steinel-cam-bridge/pkg/events"
)

// Camera-side settings and diagnostics of the L 625 (Wi-Fi signal, SD card,
// firmware, timestamp overlay, speaker volume, detection mode). They are only
// announced for cameras whose driver provides the matching callbacks.

// osdPositionNames maps the camera's timestamp overlay positions to the
// option names shown in Home Assistant. Position 0 means "off".
var osdPositionNames = map[int]string{
	1: "Oben links",
	3: "Oben rechts",
	2: "Unten links",
	4: "Unten rechts",
}

// osdPositionOptions is the option list of the position select, in display order.
var osdPositionOptions = []string{"Oben links", "Oben rechts", "Unten links", "Unten rechts"}

const defaultOSDPosition = 1

// detectModeNames maps the camera's detection modes to option names.
var detectModeNames = map[string]string{
	"ai":  "KI-Erkennung",
	"off": "Aus",
}

var detectModeOptions = []string{"KI-Erkennung", "Aus"}

func osdPositionFromName(name string) (int, bool) {
	for pos, n := range osdPositionNames {
		if strings.EqualFold(n, strings.TrimSpace(name)) {
			return pos, true
		}
	}
	return 0, false
}

func detectModeFromName(name string) (string, bool) {
	for mode, n := range detectModeNames {
		if strings.EqualFold(n, strings.TrimSpace(name)) || strings.EqualFold(mode, strings.TrimSpace(name)) {
			return mode, true
		}
	}
	return "", false
}

// rememberOSDPosition stores the last visible timestamp position for the switch.
func (c *Client) rememberOSDPosition(baseTopic string, position int) {
	if _, ok := osdPositionNames[position]; !ok {
		return
	}
	c.lowlightMu.Lock()
	defer c.lowlightMu.Unlock()
	if c.osdLast == nil {
		c.osdLast = make(map[string]int)
	}
	c.osdLast[baseTopic] = position
}

// lastOSDPosition returns the position the timestamp switch turns on with.
func (c *Client) lastOSDPosition(baseTopic string) int {
	c.lowlightMu.Lock()
	defer c.lowlightMu.Unlock()
	if p, ok := c.osdLast[baseTopic]; ok {
		return p
	}
	return defaultOSDPosition
}

// publishCameraSettingsDiscovery announces the camera-side entities.
func (c *Client) publishCameraSettingsDiscovery(dev *CameraDevice, publishEntity func(component, objectID string, config map[string]interface{})) {
	cb := dev.Callbacks
	if !cb.CameraInfo {
		return
	}
	topic := func(sub string) string { return fmt.Sprintf("%s/%s", dev.baseTopic, sub) }

	publishEntity("sensor", "wifi_signal", map[string]interface{}{
		"name":                "WLAN-Signal",
		"unit_of_measurement": "%",
		"state_class":         "measurement",
		"icon":                "mdi:wifi",
		"entity_category":     "diagnostic",
		"state_topic":         topic("wifi_signal/state"),
	})
	publishEntity("sensor", "sd_free", map[string]interface{}{
		"name":                        "SD-Karte frei",
		"unit_of_measurement":         "GB",
		"device_class":                "data_size",
		"state_class":                 "measurement",
		"suggested_display_precision": 1,
		"icon":                        "mdi:micro-sd",
		"entity_category":             "diagnostic",
		"state_topic":                 topic("sd_free/state"),
	})
	publishEntity("sensor", "sd_used", map[string]interface{}{
		"name":                        "SD-Karte belegt",
		"unit_of_measurement":         "GB",
		"device_class":                "data_size",
		"state_class":                 "measurement",
		"suggested_display_precision": 2,
		"icon":                        "mdi:micro-sd",
		"entity_category":             "diagnostic",
		"state_topic":                 topic("sd_used/state"),
	})
	publishEntity("sensor", "sd_total", map[string]interface{}{
		"name":                        "SD-Karte Größe",
		"unit_of_measurement":         "GB",
		"device_class":                "data_size",
		"suggested_display_precision": 0,
		"icon":                        "mdi:micro-sd",
		"entity_category":             "diagnostic",
		"state_topic":                 topic("sd_total/state"),
	})
	publishEntity("sensor", "sd_status", map[string]interface{}{
		"name":            "SD-Karte Status",
		"icon":            "mdi:micro-sd",
		"entity_category": "diagnostic",
		"state_topic":     topic("sd_status/state"),
	})
	publishEntity("sensor", "firmware", map[string]interface{}{
		"name":            "Firmware",
		"icon":            "mdi:chip",
		"entity_category": "diagnostic",
		"state_topic":     topic("firmware/state"),
	})

	if cb.SetTimestampPosition != nil {
		publishEntity("switch", "osd_switch", map[string]interface{}{
			"name":            "Zeitstempel im Bild",
			"icon":            "mdi:clock-digital",
			"entity_category": "config",
			"payload_on":      "ON",
			"payload_off":     "OFF",
			"state_topic":     topic("osd_switch/state"),
			"command_topic":   topic("osd_switch/set"),
		})
		publishEntity("select", "osd_position", map[string]interface{}{
			"name":            "Zeitstempel Position",
			"options":         osdPositionOptions,
			"icon":            "mdi:image-move",
			"entity_category": "config",
			"state_topic":     topic("osd_position/state"),
			"command_topic":   topic("osd_position/set"),
		})
	}
	if cb.SetSpeakerVolume != nil {
		publishEntity("number", "volume", map[string]interface{}{
			"name":                "Lautsprecher Lautstärke",
			"min":                 0,
			"max":                 100,
			"step":                5,
			"unit_of_measurement": "%",
			"icon":                "mdi:volume-high",
			"entity_category":     "config",
			"state_topic":         topic("volume/state"),
			"command_topic":       topic("volume/set"),
		})
	}
	if cb.SetDetectMode != nil {
		publishEntity("select", "detect_mode", map[string]interface{}{
			"name":            "Kamera-Erkennung",
			"options":         detectModeOptions,
			"icon":            "mdi:account-search",
			"entity_category": "config",
			"state_topic":     topic("detect_mode/state"),
			"command_topic":   topic("detect_mode/set"),
		})
	}
}

// sdGigabytes converts the camera's SD card figures to gigabytes (10^9 bytes).
// The camera names no unit; its total of 122000 for a 128 GB card (128 * 10^9
// bytes = 122070 MiB) shows that it counts in mebibytes.
func sdGigabytes(mebibytes int) float64 {
	if mebibytes < 0 {
		mebibytes = 0
	}
	return float64(mebibytes) * 1048576 / 1e9
}

// publishCameraSettingsState publishes the camera-side values that are known.
func (c *Client) publishCameraSettingsState(baseTopic string, st events.DeviceStatus, pub func(subTopic, val string)) {
	if st.WifiSignal > 0 {
		pub("wifi_signal/state", strconv.Itoa(st.WifiSignal))
	}
	if st.SDTotalMB > 0 {
		pub("sd_free/state", strconv.FormatFloat(sdGigabytes(st.SDFreeMB), 'f', 1, 64))
		pub("sd_used/state", strconv.FormatFloat(sdGigabytes(st.SDTotalMB-st.SDFreeMB), 'f', 2, 64))
		pub("sd_total/state", strconv.FormatFloat(sdGigabytes(st.SDTotalMB), 'f', 1, 64))
	}
	if st.SDStatus != "" {
		pub("sd_status/state", st.SDStatus)
	}
	if st.FirmwareVer != "" {
		pub("firmware/state", st.FirmwareVer)
	}
	if st.OSDPosition >= 0 {
		if name, ok := osdPositionNames[st.OSDPosition]; ok {
			c.rememberOSDPosition(baseTopic, st.OSDPosition)
			pub("osd_switch/state", "ON")
			pub("osd_position/state", name)
		} else {
			pub("osd_switch/state", "OFF")
			pub("osd_position/state", osdPositionNames[c.lastOSDPosition(baseTopic)])
		}
	}
	if st.SpeakerVolume >= 0 {
		pub("volume/state", strconv.Itoa(st.SpeakerVolume))
	}
	if name, ok := detectModeNames[st.DetectMode]; ok {
		pub("detect_mode/state", name)
	}
}

// handleCameraSettingsCommand processes the command topics of the camera-side
// entities. It returns false if the topic is not one of them.
func (c *Client) handleCameraSettingsCommand(topic, payload, baseTopic string, cb Callbacks) bool {
	switch {
	case strings.HasSuffix(topic, "/osd_switch/set"):
		position := 0
		if strings.EqualFold(payload, "ON") || payload == "1" || strings.EqualFold(payload, "true") {
			position = c.lastOSDPosition(baseTopic)
		}
		if cb.SetTimestampPosition != nil {
			_ = cb.SetTimestampPosition(position)
		}
		if position == 0 {
			c.publishOptimistic(baseTopic, "osd_switch/state", "OFF")
		} else {
			c.publishOptimistic(baseTopic, "osd_switch/state", "ON")
			c.publishOptimistic(baseTopic, "osd_position/state", osdPositionNames[position])
		}

	case strings.HasSuffix(topic, "/osd_position/set"):
		position, ok := osdPositionFromName(payload)
		if !ok {
			return true
		}
		c.rememberOSDPosition(baseTopic, position)
		if cb.SetTimestampPosition != nil {
			_ = cb.SetTimestampPosition(position)
		}
		c.publishOptimistic(baseTopic, "osd_position/state", osdPositionNames[position])
		c.publishOptimistic(baseTopic, "osd_switch/state", "ON")

	case strings.HasSuffix(topic, "/volume/set"):
		val, err := strconv.ParseFloat(payload, 64)
		if err != nil {
			return true
		}
		vol := int(val + 0.5)
		if vol < 0 {
			vol = 0
		} else if vol > 100 {
			vol = 100
		}
		if cb.SetSpeakerVolume != nil {
			_ = cb.SetSpeakerVolume(vol)
		}
		c.publishOptimistic(baseTopic, "volume/state", strconv.Itoa(vol))

	case strings.HasSuffix(topic, "/detect_mode/set"):
		mode, ok := detectModeFromName(payload)
		if !ok {
			return true
		}
		if cb.SetDetectMode != nil {
			_ = cb.SetDetectMode(mode)
		}
		c.publishOptimistic(baseTopic, "detect_mode/state", detectModeNames[mode])

	default:
		return false
	}
	return true
}
