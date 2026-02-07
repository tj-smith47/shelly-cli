package config

import "fmt"

// =============================================================================
// Package-level Link Functions (delegate to default manager)
// =============================================================================

// SetLink creates or updates a parent-child power link.
func SetLink(childDevice, parentDevice string, switchID int) error {
	return getDefaultManager().SetLink(childDevice, parentDevice, switchID)
}

// DeleteLink removes a link by child device name.
func DeleteLink(childDevice string) error {
	return getDefaultManager().DeleteLink(childDevice)
}

// GetLink returns a link by child device name.
func GetLink(childDevice string) (Link, bool) {
	return getDefaultManager().GetLink(childDevice)
}

// ListLinks returns all links.
func ListLinks() map[string]Link {
	return getDefaultManager().ListLinks()
}

// GetLinkedChildren returns all child device names linked to the given parent.
func GetLinkedChildren(parentDevice string) []string {
	return getDefaultManager().GetLinkedChildren(parentDevice)
}

// =============================================================================
// Manager Link Methods
// =============================================================================

// SetLink creates or updates a parent-child power link.
// Validates that both devices exist, rejects self-linking and chain linking.
func (m *Manager) SetLink(childDevice, parentDevice string, switchID int) error {
	// Reject self-linking
	childKey := NormalizeDeviceName(childDevice)
	parentKey := NormalizeDeviceName(parentDevice)
	if childKey == parentKey {
		return fmt.Errorf("cannot link device to itself")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate child device exists
	if _, ok := m.config.Devices[childKey]; !ok {
		if _, ok := m.config.Devices[childDevice]; !ok {
			return fmt.Errorf("child device %q not found", childDevice)
		}
		childKey = childDevice
	}

	// Validate parent device exists
	if _, ok := m.config.Devices[parentKey]; !ok {
		if _, ok := m.config.Devices[parentDevice]; !ok {
			return fmt.Errorf("parent device %q not found", parentDevice)
		}
		parentKey = parentDevice
	}

	// Reject chain linking: parent must not be a child in another link
	if _, isChild := m.config.Links[parentKey]; isChild {
		return fmt.Errorf("cannot chain links: %q is already a child device in another link", parentDevice)
	}

	m.config.Links[childKey] = Link{
		ParentDevice: parentKey,
		SwitchID:     switchID,
	}
	return m.saveWithoutLock()
}

// DeleteLink removes a link by child device name.
func (m *Manager) DeleteLink(childDevice string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := childDevice
	if _, ok := m.config.Links[key]; !ok {
		key = NormalizeDeviceName(childDevice)
		if _, ok := m.config.Links[key]; !ok {
			return fmt.Errorf("link for %q not found", childDevice)
		}
	}
	delete(m.config.Links, key)
	return m.saveWithoutLock()
}

// GetLink returns a link by child device name.
func (m *Manager) GetLink(childDevice string) (Link, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if link, ok := m.config.Links[childDevice]; ok {
		return link, true
	}
	link, ok := m.config.Links[NormalizeDeviceName(childDevice)]
	return link, ok
}

// ListLinks returns all links.
func (m *Manager) ListLinks() map[string]Link {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]Link, len(m.config.Links))
	for k, v := range m.config.Links {
		result[k] = v
	}
	return result
}

// GetLinkedChildren returns all child device names linked to the given parent.
func (m *Manager) GetLinkedChildren(parentDevice string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	parentKey := NormalizeDeviceName(parentDevice)
	var children []string
	for child, link := range m.config.Links {
		if link.ParentDevice == parentKey || link.ParentDevice == parentDevice {
			children = append(children, child)
		}
	}
	return children
}
