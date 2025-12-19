package config

import "time"

// Alert represents a monitoring alert configuration.
type Alert struct {
	Name         string `mapstructure:"name" json:"name" yaml:"name"`
	Description  string `mapstructure:"description,omitempty" json:"description,omitempty" yaml:"description,omitempty"`
	Device       string `mapstructure:"device" json:"device" yaml:"device"`
	Condition    string `mapstructure:"condition" json:"condition" yaml:"condition"` // e.g., "offline", "power>100", "temperature>30"
	Action       string `mapstructure:"action" json:"action" yaml:"action"`          // e.g., "notify", "webhook:http://...", "command:..."
	Enabled      bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	SnoozedUntil string `mapstructure:"snoozed_until,omitempty" json:"snoozed_until,omitempty" yaml:"snoozed_until,omitempty"`
	CreatedAt    string `mapstructure:"created_at" json:"created_at" yaml:"created_at"`
}

// IsSnoozed returns true if the alert is currently snoozed.
func (a Alert) IsSnoozed() bool {
	if a.SnoozedUntil == "" {
		return false
	}

	snoozedUntil, err := time.Parse(time.RFC3339, a.SnoozedUntil)
	if err != nil {
		return false
	}

	return time.Now().Before(snoozedUntil)
}

// Package-level functions delegate to the default manager.

// CreateAlert creates a new alert.
func CreateAlert(name, description, device, condition, action string, enabled bool) error {
	return getDefaultManager().CreateAlert(name, description, device, condition, action, enabled)
}

// DeleteAlert removes an alert.
func DeleteAlert(name string) error {
	return getDefaultManager().DeleteAlert(name)
}

// GetAlert returns an alert by name.
func GetAlert(name string) (Alert, bool) {
	return getDefaultManager().GetAlert(name)
}

// ListAlerts returns all alerts.
func ListAlerts() map[string]Alert {
	return getDefaultManager().ListAlerts()
}

// UpdateAlert updates an alert.
func UpdateAlert(name string, enabled *bool, snoozedUntil string) error {
	return getDefaultManager().UpdateAlert(name, enabled, snoozedUntil)
}
