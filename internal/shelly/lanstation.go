package shelly

import (
	"context"
	"encoding/json"
	"fmt"

	shellybackup "github.com/tj-smith47/shelly-go/backup"
	"github.com/tj-smith47/shelly-go/gen1"
	"github.com/tj-smith47/shelly-go/gen2/components"
	"github.com/tj-smith47/shelly-go/reprovision"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

// StationKey says what a LAN restore writes as a WiFi station's passphrase.
type StationKey int

// Station key outcomes of PlanLANStation and PlanAPStation.
const (
	// StationKeySkipped: --skip-network, no station is written.
	StationKeySkipped StationKey = iota
	// StationKeyAbsent: neither the backup nor the flags name a network.
	StationKeyAbsent
	// StationKeyDisabled: the backup's station is disabled; it is written
	// disabled and without a key.
	StationKeyDisabled
	// StationKeyKept: the device stays on its network and keeps its key.
	StationKeyKept
	// StationKeyGiven: the key passed with --password is written.
	StationKeyGiven
	// StationKeyBackup: the key the (Gen1) backup holds is written.
	StationKeyBackup
	// StationKeyHost: the key stored on this host is written.
	StationKeyHost
	// StationKeyBackupOpen: the backup records the network as open, and it is
	// joined as one.
	StationKeyBackupOpen
	// StationKeyOpen: the network is joined as an open one (--open).
	StationKeyOpen
	// StationKeyLeftOut: the secondary station's network changed and no
	// passphrase for it is known, so it is left out of the write.
	StationKeyLeftOut
	// StationKeyUnread: the device's stations could not be read, so the
	// secondary station is left out of the write.
	StationKeyUnread
)

// LANStationPlan is the station write a LAN restore performs.
type LANStationPlan struct {
	SSID string
	Key  StationKey
	// Address describes the IPv4 addressing the write gives the station, such
	// as "static address 10.0.0.5 (gateway 10.0.0.1, netmask 255.255.255.0)
	// from the backup" or "DHCP"; empty when the write leaves it unchanged.
	Address string
	// Secondary is the plan for the backup's secondary station (sta1); nil
	// when the backup has none that names a network.
	Secondary *LANStationPlan

	write backup.Station1Write
}

// String describes the plan for a dry-run or progress line.
func (p LANStationPlan) String() string {
	line := describeStation("WiFi station", p)
	if p.Secondary != nil {
		line += "; " + describeStation("secondary WiFi station", *p.Secondary)
	}
	return line
}

func describeStation(label string, p LANStationPlan) string {
	line := describeStationKey(label, p)
	if k := p.Key; k == StationKeySkipped || k == StationKeyAbsent || k == StationKeyLeftOut || k == StationKeyUnread {
		return line
	}
	if p.Address == "" {
		return line + "; address unchanged"
	}
	return line + "; " + p.Address
}

func describeStationKey(label string, p LANStationPlan) string {
	switch p.Key {
	case StationKeySkipped:
		return label + ": not written (--skip-network)"
	case StationKeyAbsent:
		return label + ": not written (the backup names no network)"
	case StationKeyDisabled:
		return fmt.Sprintf("%s %q: written disabled, as the backup records it, without a password", label, p.SSID)
	case StationKeyKept:
		return fmt.Sprintf("%s %q: same network as the device, which keeps its current password", label, p.SSID)
	case StationKeyGiven:
		return fmt.Sprintf("%s %q: joined with the password from --password", label, p.SSID)
	case StationKeyBackup:
		return fmt.Sprintf("%s %q: joined with the password the backup holds", label, p.SSID)
	case StationKeyHost:
		return fmt.Sprintf("%s %q: joined with the password stored on this host", label, p.SSID)
	case StationKeyBackupOpen:
		return fmt.Sprintf("%s %q: joined as an open network, as the backup records it", label, p.SSID)
	case StationKeyOpen:
		return fmt.Sprintf("%s %q: joined as an open network (--open)", label, p.SSID)
	case StationKeyLeftOut:
		return fmt.Sprintf("%s %q: left out, the device has a different network there and no password "+
			"for %q was found on this host", label, p.SSID, p.SSID)
	case StationKeyUnread:
		return fmt.Sprintf("%s %q: left out, the device's WiFi stations could not be read", label, p.SSID)
	}
	return ""
}

// Station1Write returns how the restore writes the secondary station.
func (p LANStationPlan) Station1Write() backup.Station1Write {
	if p.Secondary == nil {
		return backup.Station1Write{}
	}
	return p.Secondary.write
}

// backupStation is the part of a backup's station object that decides its key.
// Gen1 spells the switch "enabled" and Gen2+ "enable".
type backupStation struct {
	SSID    string `json:"ssid"`
	Enable  *bool  `json:"enable"`
	Enabled *bool  `json:"enabled"`
	Key     string `json:"key"`
	IsOpen  bool   `json:"is_open"`
	// Gen1 spells the addressing mode "ipv4_method" and Gen2+ "ipv4mode".
	Ipv4Method string `json:"ipv4_method"`
	IPv4Mode   string `json:"ipv4mode"`
	IP         string `json:"ip"`
	Gw         string `json:"gw"`
	Mask       string `json:"mask"`
	Netmask    string `json:"netmask"`
}

func (b backupStation) mode() string {
	if b.Ipv4Method != "" {
		return b.Ipv4Method
	}
	return b.IPv4Mode
}

// address describes the addressing a station written as the backup records
// it gets.
func (b backupStation) address() string {
	switch b.mode() {
	case ipv4Static:
		mask := b.Mask
		if mask == "" {
			mask = b.Netmask
		}
		return describeStatic(b.IP, b.Gw, mask, "from the backup")
	case ipv4DHCP:
		return "DHCP"
	}
	return ""
}

// describeStatic describes a static address and where it comes from.
func describeStatic(ip, gateway, netmask, from string) string {
	return fmt.Sprintf("static address %s (gateway %s, netmask %s) %s", ip, gateway, netmask, from)
}

func (b backupStation) disabled() bool {
	return (b.Enable != nil && !*b.Enable) || (b.Enabled != nil && !*b.Enabled)
}

func parseBackupStation(raw json.RawMessage) backupStation {
	var sta backupStation
	if len(raw) > 0 && json.Unmarshal(raw, &sta) != nil {
		return backupStation{}
	}
	return sta
}

// secondaryStation returns the "sta1" object of the backup's WiFi blob, or nil
// when the blob is missing or does not parse.
func secondaryStation(bkp *backup.DeviceBackup) json.RawMessage {
	var blob struct {
		Sta1 json.RawMessage `json:"sta1"`
	}
	if len(bkp.WiFi) == 0 || json.Unmarshal(bkp.WiFi, &blob) != nil {
		return nil
	}
	return blob.Sta1
}

// namesNetwork reports whether ov sets anything about the station.
func namesNetwork(ov *backup.NetworkOverride) bool {
	return ov != nil && (ov.SSID != "" || ov.Password != "" || ov.Open || ov.StaticIP != "")
}

// PlanLANStation decides the station passphrases for restoring bkp onto
// device over the LAN, before anything is written. The backup's network and the
// flags are merged by reprovision.MergeNetwork, the rule a --to-ap restore
// applies, so both paths write the same station for the same backup and flags
// (a backup's key and open state follow the backup's SSID only). Then:
//
//   - a disabled backup station, with no flag naming a network, is written
//     disabled and without a key, and nothing is read;
//   - an open merged network is joined as one, and a merged passphrase
//     (--password, or a key the Gen1 backup holds for its own SSID) is written;
//   - otherwise the merged SSID is compared with the device's configured one:
//     on the same network no key is sent and the device keeps its own; on a
//     different one the passphrase stored on this host is used, and with none
//     it returns PassphraseError.
//
// --open, --password or --static-ip with no SSID in the flags or the backup is
// refused, since the write would have no network to apply them to.
//
// The secondary station (sta1) follows the same rule from the backup alone,
// except that a changed network with no known passphrase, or a device whose
// stations cannot be read, leaves it out of the write instead of refusing the
// restore.
//
// The returned override is the one to restore with: ov itself, or a copy
// carrying the passphrase or open state decided here.
func (s *Service) PlanLANStation(
	ctx context.Context,
	device string,
	bkp *backup.DeviceBackup,
	ov *backup.NetworkOverride,
	skipNetwork bool,
) (*backup.NetworkOverride, LANStationPlan, error) {
	if skipNetwork {
		return ov, LANStationPlan{Key: StationKeySkipped}, nil
	}
	cur := &currentStations{svc: s, device: device}
	resolved, plan, err := s.planPrimary(ctx, cur, bkp, ov)
	if err != nil || plan.Key == StationKeyAbsent {
		return resolved, plan, err
	}
	if secondary, ok := s.planSecondary(ctx, cur, bkp); ok {
		plan.Secondary = &secondary
	}
	return resolved, plan, nil
}

// planPrimary decides the primary station's key, reading the device's current
// network only when the merged network carries neither a key nor open.
func (s *Service) planPrimary(
	ctx context.Context, cur *currentStations, bkp *backup.DeviceBackup, ov *backup.NetworkOverride,
) (*backup.NetworkOverride, LANStationPlan, error) {
	fromBackup := reprovision.NetworkFromBackup(bkp.Backup)
	given := ov.Network()
	joined, err := reprovision.MergeNetwork(&fromBackup, &given)
	if err != nil {
		return nil, LANStationPlan{}, backup.StaticNetworkError(err, ov)
	}
	plan := LANStationPlan{SSID: joined.SSID, Address: primaryAddress(bkp, given, joined)}
	resolved := backup.NetworkOverride{}
	if ov != nil {
		resolved = *ov
	}
	switch {
	case joined.SSID == "":
		if namesNetwork(ov) {
			return nil, plan, errNoNetworkNamed()
		}
		plan.Key = StationKeyAbsent
		return ov, plan, nil
	case parseBackupStation(bkp.Station()).disabled() && !namesNetwork(ov):
		plan.Key = StationKeyDisabled
		return ov, plan, nil
	}
	if key, ok := stationKeyFromMerge(&given, &joined); ok {
		plan.Key = key
		if key != StationKeyBackupOpen && key != StationKeyBackup {
			return ov, plan, nil
		}
		// The key or open state came from the backup, so the restore needs it
		// in the override it writes.
		resolved.Open = key == StationKeyBackupOpen
		resolved.Password = joined.Password
		return &resolved, plan, nil
	}
	current, _, err := cur.read(ctx)
	if err != nil {
		return nil, plan, err
	}
	if current == joined.SSID {
		plan.Key = StationKeyKept
		return ov, plan, nil
	}
	pass, err := s.hostPassphrase(ctx, joined.SSID)
	if err != nil {
		return nil, plan, err
	}
	plan.Key = StationKeyHost
	resolved.Password = pass
	return &resolved, plan, nil
}

// errNoNetworkNamed refuses network flags when neither they nor the backup
// name a network to apply them to.
func errNoNetworkNamed() error {
	return fmt.Errorf("%w: --open, --password and --static-ip need a network, and the backup "+
		"names none; pass --ssid", types.ErrInvalidParam)
}

// stationKeyFromMerge decides the station key a restore writes from the flags
// (given) and the network merged with the backup's (joined). ok is false when
// neither carries a key nor an open state, so the key has to come from the
// device or this host.
func stationKeyFromMerge(given, joined *reprovision.Network) (key StationKey, ok bool) {
	switch {
	case given.Open:
		return StationKeyOpen, true
	case given.Password != "":
		return StationKeyGiven, true
	case joined.Open:
		return StationKeyBackupOpen, true
	case joined.Password != "":
		return StationKeyBackup, true
	}
	return 0, false
}

// primaryAddress describes the addressing the primary station write gives:
// the merged static address (from --static-ip, its missing fields taken from
// the backup, or the backup's own), or the backup's DHCP mode.
func primaryAddress(bkp *backup.DeviceBackup, given, joined reprovision.Network) string {
	if joined.StaticIP == "" {
		return parseBackupStation(bkp.Station()).address()
	}
	from := "from the backup"
	if given.StaticIP != "" {
		from = "from --static-ip"
	}
	return describeStatic(joined.StaticIP, joined.Gateway, joined.Netmask, from)
}

// planSecondary plans the backup's secondary station; ok is false when the
// backup has none that names a network. A device whose stations cannot be
// read leaves sta1 out, since sta1 is never worth refusing a restore over.
func (s *Service) planSecondary(
	ctx context.Context, cur *currentStations, bkp *backup.DeviceBackup,
) (plan LANStationPlan, ok bool) {
	sta1 := parseBackupStation(secondaryStation(bkp))
	if sta1.SSID == "" {
		return plan, false
	}
	plan.SSID = sta1.SSID
	plan.Address = sta1.address()
	switch {
	case sta1.disabled():
		plan.Key = StationKeyDisabled
		return plan, true
	case sta1.IsOpen:
		plan.Key = StationKeyBackupOpen
		plan.write = backup.Station1Write{Open: true}
		return plan, true
	case sta1.Key != "":
		plan.Key = StationKeyBackup
		return plan, true
	}
	_, current, err := cur.read(ctx)
	if err != nil {
		plan.Key = StationKeyUnread
		plan.write = backup.Station1Write{Omit: true, SSID: sta1.SSID, Unread: true}
		return plan, true
	}
	if current == sta1.SSID {
		plan.Key = StationKeyKept
		return plan, true
	}
	if pass, lookupErr := s.hostPassphrase(ctx, sta1.SSID); lookupErr == nil {
		plan.Key = StationKeyHost
		plan.write = backup.Station1Write{Password: pass}
		return plan, true
	}
	plan.Key = StationKeyLeftOut
	plan.write = backup.Station1Write{Omit: true, SSID: sta1.SSID}
	return plan, true
}

// currentStations reads the device's configured station SSIDs once, on the
// first call that needs them.
type currentStations struct {
	svc        *Service
	device     string
	done       bool
	sta, sta1  string
	readFailed error
}

func (c *currentStations) read(ctx context.Context) (sta, sta1 string, err error) {
	if !c.done {
		c.done = true
		st, readErr := c.svc.deviceStations(ctx, c.device)
		c.sta, c.sta1 = st.primary.SSID, st.secondary.SSID
		if readErr != nil {
			c.readFailed = fmt.Errorf("read the device's WiFi station: %w", readErr)
		}
	}
	return c.sta, c.sta1, c.readFailed
}

// deviceStation is a station as the device has it configured.
type deviceStation struct {
	SSID   string
	Static shellybackup.StaticNetwork
}

// deviceStationPair is a device's primary and secondary station.
type deviceStationPair struct {
	primary, secondary deviceStation
}

// deviceStations returns the device's configured stations. The primary SSID
// stays set while the device runs on its secondary station. A station the
// device does not report comes back empty.
func (s *Service) deviceStations(ctx context.Context, device string) (deviceStationPair, error) {
	var out deviceStationPair
	err := s.withGenAwareAction(ctx, device,
		func(conn *client.Gen1Client) error {
			settings, err := conn.GetSettings(ctx)
			if err != nil {
				return err
			}
			out.primary = gen1Station(settings.WiFiSta)
			out.secondary = gen1Station(settings.WiFiSta1)
			return nil
		},
		func(conn *client.Client) error {
			cfg, err := components.NewWiFi(conn.RPCClient()).GetConfig(ctx)
			if err != nil {
				return err
			}
			out.primary = gen2Station(cfg.STA)
			out.secondary = gen2Station(cfg.STA1)
			return nil
		})
	return out, err
}

const (
	ipv4Static = "static"
	ipv4DHCP   = "dhcp"
)

func gen1Station(sta *gen1.WiFiStaSettings) deviceStation {
	if sta == nil {
		return deviceStation{}
	}
	st := deviceStation{SSID: sta.SSID}
	if sta.Ipv4Method == ipv4Static {
		st.Static = shellybackup.StaticNetwork{IP: sta.IP, Gateway: sta.Gw, Netmask: sta.Mask, DNS: sta.DNS}
	}
	return st
}

func gen2Station(sta *components.WiFiStationConfig) deviceStation {
	if sta == nil {
		return deviceStation{}
	}
	st := deviceStation{SSID: deref(sta.SSID)}
	if deref(sta.IPv4Mode) == ipv4Static {
		st.Static = shellybackup.StaticNetwork{
			IP: deref(sta.IP), Gateway: deref(sta.GW), Netmask: deref(sta.Netmask), DNS: deref(sta.Nameserver),
		}
	}
	return st
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
