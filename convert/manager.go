package convert

import (
	"github.com/evan-buss/opds-proxy/internal/device"
	"github.com/evan-buss/opds-proxy/internal/formats"
)

type ConverterManager struct {
	converters map[device.DeviceType]Converter
}

func NewConverterManager() *ConverterManager {
	return &ConverterManager{
		converters: map[device.DeviceType]Converter{
			device.DeviceKindle: &MobiConverter{},
			device.DeviceKobo:   &KepubConverter{},
		},
	}
}

func (cm *ConverterManager) GetConverterForDevice(deviceType device.DeviceType, format formats.Format) Converter {
	if converter, exists := cm.converters[deviceType]; exists && converter.Available() && converter.HandlesInputFormat(format) {
		return converter
	}
	return nil
}

func (cm *ConverterManager) RegisterConverter(deviceType device.DeviceType, converter Converter) {
	cm.converters[deviceType] = converter
}
