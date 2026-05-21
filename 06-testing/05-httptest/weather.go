// 05-httptest — test HTTP clients against a real (in-process) server.
//
// `httptest.NewServer` boots an HTTP server on a random port with a handler
// you supply. Your client code hits it as if it were the real upstream — same
// network stack, same parsing, same connection reuse. No mocking of the
// `http.Client`. No interface acrobatics. Just point the client at `srv.URL`.
package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Forecast is the decoded response shape.
type Forecast struct {
	City string  `json:"city"`
	TempC float64 `json:"temp_c"`
}

// Client talks to a weather API. baseURL is injected so tests can point it
// at `httptest.NewServer(...).URL` instead of a real upstream.
type Client struct {
	HTTP    *http.Client
	BaseURL string
}

// Get fetches the forecast for the given city.
// Returns an error for transport failures or non-2xx responses.
func (c *Client) Get(city string) (Forecast, error) {
	url := fmt.Sprintf("%s/forecast?city=%s", c.BaseURL, city)
	resp, err := c.HTTP.Get(url)
	if err != nil {
		return Forecast{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Forecast{}, fmt.Errorf("upstream returned %d", resp.StatusCode)
	}

	var f Forecast
	if err := json.NewDecoder(resp.Body).Decode(&f); err != nil {
		return Forecast{}, fmt.Errorf("decode: %w", err)
	}
	return f, nil
}
