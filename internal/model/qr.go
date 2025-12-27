package model

// DeviceQRInfo holds device information for QR code generation.
type DeviceQRInfo struct {
	Device    string `json:"device"`
	IP        string `json:"ip"`
	MAC       string `json:"mac,omitempty"`
	Model     string `json:"model,omitempty"`
	Firmware  string `json:"firmware,omitempty"`
	WebURL    string `json:"web_url"`
	WiFiSSID  string `json:"wifi_ssid,omitempty"`
	QRContent string `json:"qr_content"`
}
