// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// AuditDevice performs a security audit on a device and returns the results.
// Only the checks selected in checks run; reachability is always tested
// because every check needs the device to answer.
func (s *Service) AuditDevice(ctx context.Context, identifier string, checks model.AuditChecks) *model.AuditResult {
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

	info, err := s.probeDevice(ctx, identifier)
	if err != nil {
		result.Reachable = false
		result.Issues = append(result.Issues, "Device unreachable")
		return result
	}
	result.Reachable = true

	if checks.Auth {
		result.AuthStatus = &model.AuthAudit{
			AuthEnabled: info.AuthEn,
		}
		if !info.AuthEn {
			result.Issues = append(result.Issues, "Authentication is DISABLED - device is unprotected")
		} else {
			result.InfoItems = append(result.InfoItems, "Authentication enabled")
		}
	}

	if checks.Cloud {
		s.auditCloud(ctx, identifier, info.AuthEn, result)
	}
	if checks.Firmware {
		s.auditFirmware(ctx, identifier, result)
	}

	return result
}

// auditCloud records whether the device is connected to Shelly Cloud. A cloud
// connection on a device without a password is an issue whether or not the
// auth check was selected, because it is the cloud exposure being reported.
func (s *Service) auditCloud(ctx context.Context, identifier string, authEnabled bool, result *model.AuditResult) {
	cloudStatus, err := s.GetCloudStatus(ctx, identifier)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not check cloud status: %v", err))
	} else {
		result.CloudAudit = &model.CloudAudit{
			Connected: cloudStatus.Connected,
		}
		switch {
		case cloudStatus.Connected && !authEnabled:
			result.Issues = append(result.Issues, "Cloud connected but NO AUTH - exposed to internet!")
		case cloudStatus.Connected:
			result.InfoItems = append(result.InfoItems, "Cloud connected (with auth)")
		default:
			result.InfoItems = append(result.InfoItems, "Cloud not connected (local only)")
		}
	}
}

// auditFirmware records the installed firmware and whether an update is available.
func (s *Service) auditFirmware(ctx context.Context, identifier string, result *model.AuditResult) {
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
}
