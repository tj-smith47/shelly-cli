package system

// CommonTimezones is a list of common IANA timezone identifiers.
// Shelly devices use IANA timezone format (e.g., "America/New_York").
var CommonTimezones = []string{
	// UTC
	"UTC",

	// Americas
	"America/New_York",
	"America/Chicago",
	"America/Denver",
	"America/Los_Angeles",
	"America/Anchorage",
	"America/Honolulu",
	"America/Phoenix",
	"America/Toronto",
	"America/Vancouver",
	"America/Mexico_City",
	"America/Bogota",
	"America/Lima",
	"America/Santiago",
	"America/Buenos_Aires",
	"America/Sao_Paulo",

	// Europe
	"Europe/London",
	"Europe/Dublin",
	"Europe/Paris",
	"Europe/Berlin",
	"Europe/Amsterdam",
	"Europe/Brussels",
	"Europe/Vienna",
	"Europe/Zurich",
	"Europe/Rome",
	"Europe/Madrid",
	"Europe/Lisbon",
	"Europe/Warsaw",
	"Europe/Prague",
	"Europe/Budapest",
	"Europe/Stockholm",
	"Europe/Oslo",
	"Europe/Copenhagen",
	"Europe/Helsinki",
	"Europe/Athens",
	"Europe/Istanbul",
	"Europe/Moscow",
	"Europe/Kiev",
	"Europe/Bucharest",
	"Europe/Sofia",

	// Asia
	"Asia/Tokyo",
	"Asia/Seoul",
	"Asia/Shanghai",
	"Asia/Hong_Kong",
	"Asia/Taipei",
	"Asia/Singapore",
	"Asia/Bangkok",
	"Asia/Ho_Chi_Minh",
	"Asia/Jakarta",
	"Asia/Manila",
	"Asia/Kuala_Lumpur",
	"Asia/Kolkata",
	"Asia/Mumbai",
	"Asia/Karachi",
	"Asia/Dubai",
	"Asia/Riyadh",
	"Asia/Jerusalem",
	"Asia/Tehran",

	// Oceania
	"Australia/Sydney",
	"Australia/Melbourne",
	"Australia/Brisbane",
	"Australia/Perth",
	"Australia/Adelaide",
	"Pacific/Auckland",
	"Pacific/Fiji",

	// Africa
	"Africa/Cairo",
	"Africa/Johannesburg",
	"Africa/Lagos",
	"Africa/Nairobi",
	"Africa/Casablanca",
}
