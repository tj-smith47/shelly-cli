// Package tasmota provides HTTP client and types for Tasmota device communication.
package tasmota

// StatusAll represents the response from Status 0 (all status info).
type StatusAll struct {
	Status    StatusInfo `json:"Status"`
	StatusPRM StatusPRM  `json:"StatusPRM"`
	StatusFWR StatusFWR  `json:"StatusFWR"`
	StatusLOG StatusLOG  `json:"StatusLOG"`
	StatusMEM StatusMEM  `json:"StatusMEM"`
	StatusNET StatusNET  `json:"StatusNET"`
	StatusMQT StatusMQT  `json:"StatusMQT"`
	StatusTIM StatusTIM  `json:"StatusTIM"`
	StatusSNS StatusSNS  `json:"StatusSNS"`
	StatusSTS StatusSTS  `json:"StatusSTS"`
}

// StatusInfo contains basic device info from Status 1.
type StatusInfo struct {
	Module       int      `json:"Module"`
	DeviceName   string   `json:"DeviceName"`
	FriendlyName []string `json:"FriendlyName"`
	Topic        string   `json:"Topic"`
	ButtonTopic  string   `json:"ButtonTopic"`
	Power        int      `json:"Power"`
	PowerOnState int      `json:"PowerOnState"`
	LedState     int      `json:"LedState"`
	LedMask      string   `json:"LedMask"`
	SaveData     int      `json:"SaveData"`
	SaveState    int      `json:"SaveState"`
	SwitchTopic  string   `json:"SwitchTopic"`
	SwitchMode   []int    `json:"SwitchMode"`
	ButtonRetain int      `json:"ButtonRetain"`
	SwitchRetain int      `json:"SwitchRetain"`
	SensorRetain int      `json:"SensorRetain"`
	PowerRetain  int      `json:"PowerRetain"`
}

// StatusPRM contains program parameters from Status 2.
type StatusPRM struct {
	Baudrate    int    `json:"Baudrate"`
	SerialConf  string `json:"SerialConfig"`
	GroupTopic  string `json:"GroupTopic"`
	OtaURL      string `json:"OtaUrl"`
	RestartRsn  string `json:"RestartReason"`
	Uptime      string `json:"Uptime"`
	StartupUTC  string `json:"StartupUTC"`
	Sleep       int    `json:"Sleep"`
	CfgHolder   int    `json:"CfgHolder"`
	BootCount   int    `json:"BootCount"`
	BCResetTime string `json:"BCResetTime"`
	SaveCount   int    `json:"SaveCount"`
	SaveAddress string `json:"SaveAddress"`
}

// StatusFWR contains firmware info from Status 2.
type StatusFWR struct {
	Version       string `json:"Version"`
	BuildDateTime string `json:"BuildDateTime"`
	Boot          int    `json:"Boot"`
	Core          string `json:"Core"`
	SDK           string `json:"SDK"`
	CPUFrequency  int    `json:"CpuFrequency"`
	Hardware      string `json:"Hardware"`
	CR            string `json:"CR"`
}

// StatusLOG contains logging settings from Status 3.
type StatusLOG struct {
	SerialLog  int      `json:"SerialLog"`
	WebLog     int      `json:"WebLog"`
	MqttLog    int      `json:"MqttLog"`
	SysLog     int      `json:"SysLog"`
	LogHost    string   `json:"LogHost"`
	LogPort    int      `json:"LogPort"`
	SSId       []any    `json:"SSId"`
	TelePrd    int      `json:"TelePeriod"`
	Resolution string   `json:"Resolution"`
	SetOption  []string `json:"SetOption"`
}

// StatusMEM contains memory info from Status 4.
type StatusMEM struct {
	ProgramSize    int      `json:"ProgramSize"`
	Free           int      `json:"Free"`
	Heap           int      `json:"Heap"`
	ProgramFlashSz int      `json:"ProgramFlashSize"`
	FlashSize      int      `json:"FlashSize"`
	FlashChipID    string   `json:"FlashChipId"`
	FlashFrequency int      `json:"FlashFrequency"`
	FlashMode      int      `json:"FlashMode"`
	Features       []string `json:"Features"`
	Drivers        string   `json:"Drivers"`
	Sensors        string   `json:"Sensors"`
}

// StatusNET contains network info from Status 5.
type StatusNET struct {
	Hostname   string `json:"Hostname"`
	IPAddress  string `json:"IPAddress"`
	Gateway    string `json:"Gateway"`
	Subnetmask string `json:"Subnetmask"`
	DNSServer1 string `json:"DNSServer1"`
	DNSServer2 string `json:"DNSServer2"`
	Mac        string `json:"Mac"`
	Webserver  int    `json:"Webserver"`
	HTTPAPI    int    `json:"HTTP_API"`
	WifiConfig int    `json:"WifiConfig"`
	WifiPower  int    `json:"WifiPower"`
}

// StatusMQT contains MQTT info from Status 6.
type StatusMQT struct {
	MqttHost      string `json:"MqttHost"`
	MqttPort      int    `json:"MqttPort"`
	MqttClientMk  string `json:"MqttClientMask"`
	MqttClient    string `json:"MqttClient"`
	MqttUser      string `json:"MqttUser"`
	MqttCount     int    `json:"MqttCount"`
	MaxPacket     int    `json:"MAX_PACKET_SIZE"`
	Keepalive     int    `json:"KEEPALIVE"`
	SocketTimeout int    `json:"SOCKET_TIMEOUT"`
}

// StatusTIM contains time info from Status 7.
type StatusTIM struct {
	UTC      string `json:"UTC"`
	Local    string `json:"Local"`
	StartDST string `json:"StartDST"`
	EndDST   string `json:"EndDST"`
	Timezone any    `json:"Timezone"`
	Sunrise  string `json:"Sunrise"`
	Sunset   string `json:"Sunset"`
}

// StatusSNS contains sensor readings from Status 8.
type StatusSNS struct {
	Time   string       `json:"Time"`
	Energy *EnergyStats `json:"ENERGY,omitempty"`
	// Additional sensors can be added as needed
	DS18B20 *DS18B20 `json:"DS18B20,omitempty"`
	AM2301  *AM2301  `json:"AM2301,omitempty"`
	BME280  *BME280  `json:"BME280,omitempty"`
}

// StatusSTS contains power state from Status 10 (Tele state).
type StatusSTS struct {
	Time      string `json:"Time"`
	Uptime    string `json:"Uptime"`
	UptimeSec int    `json:"UptimeSec"`
	Heap      int    `json:"Heap"`
	SleepMode string `json:"SleepMode"`
	Sleep     int    `json:"Sleep"`
	LoadAvg   int    `json:"LoadAvg"`
	MqttCount int    `json:"MqttCount"`
	// Power states - Tasmota names them POWER, POWER1, POWER2, etc.
	Power  string  `json:"POWER,omitempty"`
	Power1 string  `json:"POWER1,omitempty"`
	Power2 string  `json:"POWER2,omitempty"`
	Power3 string  `json:"POWER3,omitempty"`
	Power4 string  `json:"POWER4,omitempty"`
	Wifi   WifiSTS `json:"Wifi"`
}

// WifiSTS contains WiFi status.
type WifiSTS struct {
	AP       int    `json:"AP"`
	SSId     string `json:"SSId"`
	BSSId    string `json:"BSSId"`
	Channel  int    `json:"Channel"`
	Mode     string `json:"Mode"`
	RSSI     int    `json:"RSSI"`
	Signal   int    `json:"Signal"`
	LinkCnt  int    `json:"LinkCount"`
	Downtime string `json:"Downtime"`
}

// EnergyStats contains energy meter readings.
type EnergyStats struct {
	TotalStartTime string  `json:"TotalStartTime"`
	Total          float64 `json:"Total"`
	Yesterday      float64 `json:"Yesterday"`
	Today          float64 `json:"Today"`
	Period         int     `json:"Period"`
	Power          float64 `json:"Power"`
	ApparentPower  float64 `json:"ApparentPower"`
	ReactivePower  float64 `json:"ReactivePower"`
	Factor         float64 `json:"Factor"`
	Voltage        float64 `json:"Voltage"`
	Current        float64 `json:"Current"`
}

// DS18B20 represents a DS18B20 temperature sensor.
type DS18B20 struct {
	ID          string  `json:"Id"`
	Temperature float64 `json:"Temperature"`
}

// AM2301 represents an AM2301/DHT21 temperature/humidity sensor.
type AM2301 struct {
	Temperature float64 `json:"Temperature"`
	Humidity    float64 `json:"Humidity"`
	DewPoint    float64 `json:"DewPoint"`
}

// BME280 represents a BME280 environmental sensor.
type BME280 struct {
	Temperature float64 `json:"Temperature"`
	Humidity    float64 `json:"Humidity"`
	DewPoint    float64 `json:"DewPoint"`
	Pressure    float64 `json:"Pressure"`
}

// PowerResponse is the response from Power commands.
type PowerResponse struct {
	Power  string `json:"POWER,omitempty"`
	Power1 string `json:"POWER1,omitempty"`
	Power2 string `json:"POWER2,omitempty"`
	Power3 string `json:"POWER3,omitempty"`
	Power4 string `json:"POWER4,omitempty"`
}

// UpgradeResponse is the response from Upgrade command.
type UpgradeResponse struct {
	Upgrade string `json:"Upgrade"`
}
