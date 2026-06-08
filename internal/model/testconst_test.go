package model

// Shared test-data constants for the model package test suite. Centralised here
// so repeated literals stay consistent across files and satisfy goconst.
const (
	testDeviceID123 = "device-123"
	testDeviceID456 = "device-456"
	testDevice1     = "device1"
	testDevice2     = "device2"

	testStatusOnline       = "online"
	testStatusDisconnected = "disconnected"

	testSourceButton = "button"

	testNameTest      = "test"
	testUserAdmin     = "admin"
	testPasswordValue = "secret123"

	testLivingRoomSwitch = "Living Room Switch"
	testKitchenSwitch    = "Kitchen Switch"
	testShellyPlus1PM    = "Shelly Plus 1PM"
	testPlus1PM          = "Plus 1PM"
	testModelSHSWPM      = "SHSW-PM"

	testIP50 = "192.168.1.50"
	testURL  = "http://192.168.1.100"
	testSSID = "MyNetwork"

	testFW100 = "1.0.0"
	testFW110 = "1.1.0"

	testGenTasmota = "tasmota"
	testGenESPHome = "esphome"

	testSwitch1 = "switch1"

	testReportTypeInventory = "inventory"
)
