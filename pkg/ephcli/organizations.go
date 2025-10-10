// Package ephcli provides organization management methods for the ephemeralfiles API.
package ephcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ephemeralfiles/eph/pkg/dto"
)

var (
	// ErrOrganizationNotFound is returned when an organization cannot be found.
	ErrOrganizationNotFound = errors.New("organization not found")
)

// ListOrganizations retrieves all organizations for the authenticated user.
func (c *ClientEphemeralfiles) ListOrganizations() ([]dto.Organization, error) {
	url := fmt.Sprintf("%s/%s/organizations", c.endpoint, apiVersion)

	req, cancel, err := c.createRequestWithTimeout(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	resp, err := c.doRequestWithAuth(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var orgs []dto.Organization
	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return nil, fmt.Errorf("failed to decode organizations response: %w", err)
	}

	return orgs, nil
}

// GetOrganization retrieves a specific organization by ID.
func (c *ClientEphemeralfiles) GetOrganization(orgID string) (*dto.Organization, error) {
	url := fmt.Sprintf("%s/%s/organizations/%s", c.endpoint, apiVersion, orgID)

	req, cancel, err := c.createRequestWithTimeout(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	resp, err := c.doRequestWithAuth(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var org dto.Organization
	if err := json.NewDecoder(resp.Body).Decode(&org); err != nil {
		return nil, fmt.Errorf("failed to decode organization response: %w", err)
	}

	return &org, nil
}

// GetOrganizationByName retrieves an organization by name.
func (c *ClientEphemeralfiles) GetOrganizationByName(name string) (*dto.Organization, error) {
	orgs, err := c.ListOrganizations()
	if err != nil {
		return nil, err
	}

	for i := range orgs {
		if orgs[i].Name == name {
			return &orgs[i], nil
		}
	}

	return nil, ErrOrganizationNotFound
}

// GetOrganizationStorage retrieves storage information for an organization.
func (c *ClientEphemeralfiles) GetOrganizationStorage(orgID string) (*dto.OrganizationStorage, error) {
	url := fmt.Sprintf("%s/%s/organizations/%s/storage", c.endpoint, apiVersion, orgID)

	req, cancel, err := c.createRequestWithTimeout(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	resp, err := c.doRequestWithAuth(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var storage dto.OrganizationStorage
	if err := json.NewDecoder(resp.Body).Decode(&storage); err != nil {
		return nil, fmt.Errorf("failed to decode storage response: %w", err)
	}

	return &storage, nil
}

// GetOrganizationStats retrieves organization statistics.
func (c *ClientEphemeralfiles) GetOrganizationStats(orgID string) (*dto.OrganizationStats, error) {
	url := fmt.Sprintf("%s/%s/organizations/%s/stats", c.endpoint, apiVersion, orgID)

	req, cancel, err := c.createRequestWithTimeout(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	resp, err := c.doRequestWithAuth(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var stats dto.OrganizationStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode stats response: %w", err)
	}

	return &stats, nil
}
