package caddy

import (
	"context"
	"errors"

	"caddy-manager/internal/models"
)

type MockClient struct {
	IsReachable bool
	Version     string
	ConfigData  []byte
	ApplyError  error
}

func NewMockClient(reachable bool, version string) *MockClient {
	return &MockClient{
		IsReachable: reachable,
		Version:     version,
		ConfigData:  []byte(`{"apps":{}}`),
	}
}

func (m *MockClient) Health(ctx context.Context) error {
	if !m.IsReachable {
		return errors.New("mock caddy unreachable")
	}
	return nil
}

func (m *MockClient) GetVersion(ctx context.Context) (string, error) {
	if !m.IsReachable {
		return "", errors.New("mock caddy unreachable")
	}
	return m.Version, nil
}

func (m *MockClient) GetConfig(ctx context.Context) ([]byte, error) {
	if !m.IsReachable {
		return nil, errors.New("mock caddy unreachable")
	}
	return m.ConfigData, nil
}

func (m *MockClient) ApplyConfig(ctx context.Context, configBytes []byte) error {
	if !m.IsReachable {
		return errors.New("mock caddy unreachable")
	}
	if m.ApplyError != nil {
		return m.ApplyError
	}
	m.ConfigData = configBytes
	return nil
}

func (m *MockClient) GetSites(ctx context.Context) (models.SiteDiscoveryResult, error) {
	if !m.IsReachable {
		return models.SiteDiscoveryResult{}, errors.New("mock caddy unreachable")
	}
	return ParseSitesFromConfig(m.ConfigData)
}

func (m *MockClient) ApplySite(ctx context.Context, site models.Site) error {
	if !m.IsReachable {
		return errors.New("mock caddy unreachable")
	}
	return nil
}

func (m *MockClient) DeleteSite(ctx context.Context, siteID string) error {
	if !m.IsReachable {
		return errors.New("mock caddy unreachable")
	}
	return nil
}
