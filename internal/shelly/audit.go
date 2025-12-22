// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// AuditDevice performs a security audit on a device and returns the results.
func (s *Service) AuditDevice(ctx context.Context, identifier string) *model.AuditResult {
	result := &model.AuditResult{
		Device:    identifier,
		Issues:    []string{},
		Warnings:  []string{},
		InfoItems: []string{},
	}

	// Resolve device to get address
	device, err := s.resolver.Resolve(identifier)
	if err != nil {
		result.Address = identifier
	} else {
		result.Address = device.Address
	}

	// Try to ping device first
	info, err := s.DevicePing(ctx, identifier)
	if err != nil {
		result.Reachable = false
		result.Issues = append(result.Issues, "Device unreachable")
		return result
	}
	result.Reachable = true

	// Check authentication status
	result.AuthStatus = &model.AuthAudit{
		AuthEnabled: info.AuthEn,
	}
	if !info.AuthEn {
		result.Issues = append(result.Issues, "Authentication is DISABLED - device is unprotected")
	} else {
		result.InfoItems = append(result.InfoItems, "Authentication enabled")
	}

	// Check cloud status
	cloudStatus, err := s.GetCloudStatus(ctx, identifier)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not check cloud status: %v", err))
	} else {
		result.CloudAudit = &model.CloudAudit{
			Connected: cloudStatus.Connected,
		}
		switch {
		case cloudStatus.Connected && !info.AuthEn:
			result.Issues = append(result.Issues, "Cloud connected but NO AUTH - exposed to internet!")
		case cloudStatus.Connected:
			result.InfoItems = append(result.InfoItems, "Cloud connected (with auth)")
		default:
			result.InfoItems = append(result.InfoItems, "Cloud not connected (local only)")
		}
	}

	// Check firmware
	fwInfo, err := s.CheckFirmware(ctx, identifier)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not check firmware: %v", err))
	} else {
		result.FWAudit = &model.FirmwareAudit{
			Current:   fwInfo.Current,
			Available: fwInfo.Available,
			HasUpdate: fwInfo.HasUpdate,
		}
		if fwInfo.HasUpdate {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Firmware update available: %s -> %s", fwInfo.Current, fwInfo.Available))
		} else {
			result.InfoItems = append(result.InfoItems,
				fmt.Sprintf("Firmware up to date (%s)", fwInfo.Current))
		}
	}

	return result
}
