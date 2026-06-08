package term

// Shared test fixtures. goconst flags identical literals repeated across the
// package's test files; these consts keep the values in one place.
const (
	testIP10  = "192.168.1.10"
	testIP50  = "192.168.1.50"
	testIP100 = "192.168.1.100"
	testIP101 = "192.168.1.101"

	testMAC  = "AA:BB:CC:DD:EE:FF"
	testMAC2 = "11:22:33:44:55:66"

	testModel1PM = "SNSW-001P16EU"
	testModel1X  = "SNSW-001X16EU"
	testModel2PM = "SNSW-002P16EU"

	testFWVersion     = "1.0.0"
	testFWVersionNew  = "1.1.0"
	testFWVersionBeta = "1.2.0-beta1"
	testFWVersion2    = "2.0.0"
	testVersion10     = "1.0"

	testInput0      = "Input 0"
	testKitchenName = "Kitchen Light"
	testHighTemp    = "High Temp"
	testShellyPlus1 = "Shelly Plus 1"
	testMethodBLE   = "BLE"

	testCron8am  = "0 0 8 * * *"
	testCron10pm = "0 0 22 * * *"

	testDevice1 = "device1"
	testDevice2 = "device2"
	testDevice3 = "device3"
	testDevID1  = "dev-1"
	testDevID2  = "dev-2"

	testDeviceIDPlus1 = "shellyplus1-123456"
	testDeviceIDPro1  = "shellypro1pm-123456"
	testModelNamePro1 = "Shelly Pro 1PM"

	testKey   = "key"
	testKey1  = "key1"
	testKey2  = "key2"
	testKey3  = "key3"
	testValue = "value"
	testVal1  = "value1"
	testETag  = "etag123"

	testConfigSwitchName = "switch:0.name"
	testConfigDeviceName = "sys.device.name"

	testSwitchSet    = "Switch.Set"
	testSwitchToggle = "Switch.Toggle"
	testSwitch0      = "Switch 0"
	testCompSwitch   = "switch"

	testTemperature    = "temperature"
	testTempCondition  = "temperature > 30"
	testAlertName      = "Test Alert"
	testGroupName      = "Test Group"
	testTypeCustom     = "custom"
	testTypeHeat       = "heat"
	testValueConnected = "connected"
	testValueServer    = "server"
	testValueShelly    = "shelly"
	testValueOutput    = "output"
	testValueTest      = "test"
	testValueZero      = "zero"
	testWSURL          = "ws://example.com"
)
