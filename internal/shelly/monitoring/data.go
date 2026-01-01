package monitoring

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// GetEMDataRecords retrieves available time intervals with stored EMData.
func (s *Service) GetEMDataRecords(ctx context.Context, device string, id int, fromTS *int64) (*components.EMDataRecordsResult, error) {
	var result *components.EMDataRecordsResult
	err := s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		emdata := components.NewEMData(conn.RPCClient(), id)
		var err error
		result, err = emdata.GetRecords(ctx, fromTS)
		return err
	})
	return result, err
}

// GetEMDataHistory retrieves historical EMData measurements for a time range.
func (s *Service) GetEMDataHistory(ctx context.Context, device string, id int, startTS, endTS *int64) (*components.EMDataGetDataResult, error) {
	var result *components.EMDataGetDataResult
	err := s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		emdata := components.NewEMData(conn.RPCClient(), id)
		var err error
		result, err = emdata.GetData(ctx, startTS, endTS)
		return err
	})
	return result, err
}

// DeleteEMData deletes all stored historical EMData.
func (s *Service) DeleteEMData(ctx context.Context, device string, id int) error {
	return s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		emdata := components.NewEMData(conn.RPCClient(), id)
		return emdata.DeleteAllData(ctx)
	})
}

// GetEMDataCSVURL returns the HTTP URL for downloading EMData as CSV.
func (s *Service) GetEMDataCSVURL(device string, id int, startTS, endTS *int64, addKeys bool) (string, error) {
	addr, err := s.connector.Resolve(device)
	if err != nil {
		return "", err
	}
	return components.EMDataCSVURL(addr.Address, id, startTS, endTS, addKeys), nil
}

// GetEM1DataRecords retrieves available time intervals with stored EM1Data.
func (s *Service) GetEM1DataRecords(ctx context.Context, device string, id int, fromTS *int64) (*components.EM1DataRecordsResult, error) {
	var result *components.EM1DataRecordsResult
	err := s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		em1data := components.NewEM1Data(conn.RPCClient(), id)
		var err error
		result, err = em1data.GetRecords(ctx, fromTS)
		return err
	})
	return result, err
}

// GetEM1DataHistory retrieves historical EM1Data measurements for a time range.
func (s *Service) GetEM1DataHistory(ctx context.Context, device string, id int, startTS, endTS *int64) (*components.EM1DataGetDataResult, error) {
	var result *components.EM1DataGetDataResult
	err := s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		em1data := components.NewEM1Data(conn.RPCClient(), id)
		var err error
		result, err = em1data.GetData(ctx, startTS, endTS)
		return err
	})
	return result, err
}

// DeleteEM1Data deletes all stored historical EM1Data.
func (s *Service) DeleteEM1Data(ctx context.Context, device string, id int) error {
	return s.connector.WithConnection(ctx, device, func(conn *client.Client) error {
		em1data := components.NewEM1Data(conn.RPCClient(), id)
		return em1data.DeleteAllData(ctx)
	})
}

// GetEM1DataCSVURL returns the HTTP URL for downloading EM1Data as CSV.
func (s *Service) GetEM1DataCSVURL(device string, id int, startTS, endTS *int64, addKeys bool) (string, error) {
	addr, err := s.connector.Resolve(device)
	if err != nil {
		return "", err
	}
	return components.EM1DataCSVURL(addr.Address, id, startTS, endTS, addKeys), nil
}
