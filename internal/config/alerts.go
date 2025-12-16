package config

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
