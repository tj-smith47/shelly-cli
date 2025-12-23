// Package shelly provides device model name mappings.
package shelly

// modelNames maps Shelly model codes to human-readable names.
// Based on official Shelly documentation.
var modelNames = map[string]string{
	// Gen2+ Switches
	"SNSW-001X16EU":  "Shelly Plus 1",
	"SNSW-001P16EU":  "Shelly Plus 1PM",
	"SNSW-002P16EU":  "Shelly Plus 2PM",
	"SNSW-102P16EU":  "Shelly Plus 2PM (v2)",
	"SNSW-001P8EU":   "Shelly Plus 1 Mini",
	"SNPM-001PCEU16": "Shelly Plus PM Mini",

	// Gen2+ Plugs
	"SNPL-00112EU": "Shelly Plus Plug S",
	"SNPL-00116EU": "Shelly Plus Plug EU",
	"SNPL-00116US": "Shelly Plus Plug US",
	"SNPL-00116UK": "Shelly Plus Plug UK",

	// Gen2+ Dimmers
	"SNDM-00100WW": "Shelly Plus Wall Dimmer",
	"SNDM-0013US":  "Shelly Plus 0-10V Dimmer",

	// Gen2+ Energy Meters
	"SPEM-002CEBEU50": "Shelly Pro 3EM",
	"SPEM-003CEBEU":   "Shelly Pro EM-50",
	"SNEM-001PCEU":    "Shelly Plus EM",

	// Gen2+ Pro Series
	"SPSW-001XE16EU":  "Shelly Pro 1",
	"SPSW-001PE16EU":  "Shelly Pro 1PM",
	"SPSW-002PE16EU":  "Shelly Pro 2PM",
	"SPSW-102PE16EU":  "Shelly Pro 2PM (v2)",
	"SPSW-003XE16EU":  "Shelly Pro 3",
	"SPSW-004PE16EU":  "Shelly Pro 4PM",
	"SPSH-002PE16EU":  "Shelly Pro Dual Cover PM",

	// Gen2+ i4 Inputs
	"SNSN-0024X":  "Shelly Plus i4",
	"SNSN-0D24X":  "Shelly Plus i4 DC",
	"SPSN-0024X4": "Shelly Pro i4",

	// Gen2+ Covers
	"SNSC-0D4P":  "Shelly Plus Cover",

	// Gen2+ RGB/RGBW
	"SNDC-0D4P":   "Shelly Plus RGBW PM",
	"SNDM-00N100": "Shelly Plus Dimmer",

	// Gen2+ Add-ons
	"SNAD-0001A10EU": "Shelly Plus Add-on",
	"SPAD-0001A10EU": "Shelly Pro Add-on",

	// Gen2+ Sensors
	"SNSN-0031Z":   "Shelly Plus H&T",
	"SNSE-0043US":  "Shelly Plus Smoke",
	"SNSN-0043X":   "Shelly BLU Motion",
	"SNSN-0013X":   "Shelly BLU Gateway",

	// Gen1 Devices (legacy)
	"SHSW-1":    "Shelly 1",
	"SHSW-PM":   "Shelly 1PM",
	"SHSW-21":   "Shelly 2",
	"SHSW-25":   "Shelly 2.5",
	"SHPLG-1":   "Shelly Plug",
	"SHPLG-S":   "Shelly Plug S",
	"SHPLG2-1":  "Shelly Plug US",
	"SHDM-1":    "Shelly Dimmer",
	"SHDM-2":    "Shelly Dimmer 2",
	"SHRGBW2":   "Shelly RGBW2",
	"SHBDUO-1":  "Shelly Duo",
	"SHVIN-1":   "Shelly Vintage",
	"SHEM":      "Shelly EM",
	"SHEM-3":    "Shelly 3EM",
	"SHHT-1":    "Shelly H&T",
	"SHSM-01":   "Shelly Smoke",
	"SHWT-1":    "Shelly Flood",
	"SHDW-1":    "Shelly Door/Window",
	"SHDW-2":    "Shelly Door/Window 2",
	"SHGS-1":    "Shelly Gas",
	"SHMOS-01":  "Shelly Motion",
	"SHMOS-02":  "Shelly Motion 2",
	"SHBTN-1":   "Shelly Button1",
	"SHBTN-2":   "Shelly Button",
	"SHIX3-1":   "Shelly i3",
	"SHUNI-1":   "Shelly Uni",
	"SHAIR-1":   "Shelly Air",
	"SHSEN-1":   "Shelly Sense",
}

// ModelDisplayName returns a human-readable name for a Shelly model code.
// If no mapping exists, returns the original model code.
func ModelDisplayName(modelCode string) string {
	if name, ok := modelNames[modelCode]; ok {
		return name
	}
	return modelCode
}

// ModelShortName returns a shortened display name.
// Removes "Shelly " prefix for compact display.
func ModelShortName(modelCode string) string {
	name := ModelDisplayName(modelCode)
	if len(name) > 7 && name[:7] == "Shelly " {
		return name[7:]
	}
	return name
}

// ModelDisplayName returns a human-readable name for a model code via the service.
func (s *Service) ModelDisplayName(modelCode string) string {
	return ModelDisplayName(modelCode)
}

// ModelShortName returns a shortened display name via the service.
func (s *Service) ModelShortName(modelCode string) string {
	return ModelShortName(modelCode)
}
