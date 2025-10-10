// Package ephcli provides organization context management for resolving organization from flags or configuration.
package ephcli

import (
	"errors"
	"fmt"

	"github.com/ephemeralfiles/eph/pkg/config"
	"github.com/ephemeralfiles/eph/pkg/dto"
)

var (
	// ErrNoOrganizationSpecified is returned when no organization is specified via flags or config.
	ErrNoOrganizationSpecified = errors.New("no organization specified. Use --org flag or set default with 'eph org use'")
)

// OrgContext manages organization context resolution.
type OrgContext struct {
	client *ClientEphemeralfiles
	config *config.Config
}

// NewOrgContext creates a new organization context manager.
func NewOrgContext(client *ClientEphemeralfiles, cfg *config.Config) *OrgContext {
	return &OrgContext{
		client: client,
		config: cfg,
	}
}

// ResolveOrganization resolves organization from flags or config.
// Priority: --org-id flag > --org flag > config.DefaultOrganization.
func (oc *OrgContext) ResolveOrganization(orgFlag string, orgIDFlag string) (*dto.Organization, error) {
	// Priority 1: Explicit org-id flag
	if orgIDFlag != "" {
		return oc.client.GetOrganization(orgIDFlag)
	}

	// Priority 2: Explicit org name flag
	if orgFlag != "" {
		return oc.client.GetOrganizationByName(orgFlag)
	}

	// Priority 3: Config default organization
	if oc.config.DefaultOrganization != "" {
		// Try as name first, then as ID if name lookup fails
		org, err := oc.client.GetOrganizationByName(oc.config.DefaultOrganization)
		if err != nil {
			// If name lookup fails, try as ID
			org, err = oc.client.GetOrganization(oc.config.DefaultOrganization)
			if err != nil {
				return nil, fmt.Errorf("default organization '%s' not found: %w", oc.config.DefaultOrganization, err)
			}
		}
		return org, nil
	}

	// No organization specified
	return nil, ErrNoOrganizationSpecified
}

// ResolveOrganizationID gets organization ID from name or returns ID directly.
func (oc *OrgContext) ResolveOrganizationID(nameOrID string) (string, error) {
	// First try to get by name
	org, err := oc.client.GetOrganizationByName(nameOrID)
	if err == nil {
		return org.ID, nil
	}

	// If that fails, try to get by ID
	org, err = oc.client.GetOrganization(nameOrID)
	if err != nil {
		return "", fmt.Errorf("organization '%s' not found: %w", nameOrID, err)
	}

	return org.ID, nil
}
