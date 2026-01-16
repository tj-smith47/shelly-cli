package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// ComponentResponse represents a component in the GetComponents response.
type ComponentResponse struct {
	Key    string          `json:"key"`
	Status json.RawMessage `json:"status,omitempty"`
	Config json.RawMessage `json:"config,omitempty"`
}

// getComponentsResponse is the response from Shelly.GetComponents.
type getComponentsResponse struct {
	Components []ComponentResponse `json:"components"`
	Offset     int                 `json:"offset"`
	Total      int                 `json:"total"`
}

// GetComponentsAll fetches all components from the device with pagination.
// It accepts optional base params (e.g., dynamic_only, include_status, include_config).
// Returns the raw component responses for caller to process.
func (c *Client) GetComponentsAll(ctx context.Context, baseParams map[string]any) ([]ComponentResponse, error) {
	var allComps []ComponentResponse
	offset := 0

	for {
		// Build params with offset
		params := make(map[string]any)
		for k, v := range baseParams {
			params[k] = v
		}
		if offset > 0 {
			params["offset"] = offset
		}

		// Make the call
		result, err := c.rpcClient.Call(ctx, "Shelly.GetComponents", params)
		if err != nil {
			return nil, fmt.Errorf("failed to get components: %w", err)
		}

		// Parse response
		var resp getComponentsResponse
		if err := unmarshalResponse(result, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse components: %w", err)
		}

		allComps = append(allComps, resp.Components...)

		// Check if we've fetched all components
		if resp.Total == 0 || resp.Offset+len(resp.Components) >= resp.Total {
			break
		}
		offset = resp.Offset + len(resp.Components)
	}

	return allComps, nil
}
