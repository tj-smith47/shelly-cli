package shelly

// The device profile registry is filled by the init functions of these
// packages. Nothing else imports them, so without these imports every profile
// lookup (profile list, profile info, zwave info) finds no models.
import (
	_ "github.com/tj-smith47/shelly-go/profiles/blu"
	_ "github.com/tj-smith47/shelly-go/profiles/gen1"
	_ "github.com/tj-smith47/shelly-go/profiles/gen2"
	_ "github.com/tj-smith47/shelly-go/profiles/gen3"
	_ "github.com/tj-smith47/shelly-go/profiles/gen4"
	_ "github.com/tj-smith47/shelly-go/profiles/wave"
)
