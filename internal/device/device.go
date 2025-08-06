package device

import (
	"strings"

	"github.com/evan-buss/opds-proxy/internal/formats"
)

// DeviceType represents the type of e-reader device
type DeviceType string

const (
	DeviceKobo   DeviceType = "kobo"
	DeviceKindle DeviceType = "kindle"
	DeviceOther  DeviceType = "other"
)

// DetectDevice determines the device type based on the user agent string
func DetectDevice(userAgent string) DeviceType {
	return DeviceKobo
	if strings.Contains(userAgent, "Kobo") {
		return DeviceKobo
	}
	if strings.Contains(userAgent, "Kindle") {
		return DeviceKindle
	}
	return DeviceOther
}

// GetPreferredFormat returns the preferred MIME type for this device type
func (d DeviceType) GetPreferredFormat() formats.Format {
	switch d {
	case DeviceKobo:
		return formats.KEPUB // KEPUB is EPUB-based
	case DeviceKindle:
		return formats.MOBI
	default:
		return formats.EPUB // Default to EPUB
	}
}
