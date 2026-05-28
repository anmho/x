package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultLinearAPIURL = "https://api.linear.app/graphql"

type LinearIssue struct {
	ID         string
	Identifier string
	Title      string
	URL        string
	StateName  string
	StateType  string
	TeamID     string
	TeamKey    string
	TeamName   string
}

type LinearClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewLinearClientFromEnv() *LinearClient {
	return &LinearClient{
		apiKey:  strings.TrimSpace(os.Getenv("LINEAR_API_KEY")),
		baseURL: strings.TrimSpace(os.Getenv("LINEAR_API_URL")),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *LinearClient) Enabled() bool {
	return c != nil && c.apiKey != ""
}

func (c *LinearClient) GetIssue(ctx context.Context, identifier string) (*LinearIssue, error) {
	if c == nil || !c.Enabled() {
		return nil, errors.New("linear client is not configured")
	}
	identifier = strings.ToUpper(strings.TrimSpace(identifier))
	if identifier == "" {
		return nil, errors.New("issue identifier is required")
	}

	query := `
query IssueByIdentifier($identifier: String!) {
  issues(first: 1, filter: { identifier: { eq: $identifier } }) {
    nodes {
      id
      identifier
      title
      url
      state {
        name
        type
      }
      team {
        id
        key
        name
      }
    }
  }
}`

	payload := map[string]any{
		"query": query,
		"variables": map[string]string{
			"identifier": identifier,
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	baseURL := c.baseURL
	if baseURL == "" {
		baseURL = defaultLinearAPIURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Data struct {
			Issues struct {
				Nodes []struct {
					ID         string `json:"id"`
					Identifier string `json:"identifier"`
					Title      string `json:"title"`
					URL        string `json:"url"`
					State      struct {
						Name string `json:"name"`
						Type string `json:"type"`
					} `json:"state"`
					Team struct {
						ID   string `json:"id"`
						Key  string `json:"key"`
						Name string `json:"name"`
					} `json:"team"`
				} `json:"nodes"`
			} `json:"issues"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		if len(data.Errors) > 0 {
			return nil, fmt.Errorf("linear request failed: %s", data.Errors[0].Message)
		}
		return nil, fmt.Errorf("linear request failed: status %d", resp.StatusCode)
	}
	if len(data.Errors) > 0 {
		return nil, fmt.Errorf("linear request failed: %s", data.Errors[0].Message)
	}
	if len(data.Data.Issues.Nodes) == 0 {
		return nil, fmt.Errorf("linear issue %q not found", identifier)
	}
	node := data.Data.Issues.Nodes[0]
	return &LinearIssue{
		ID:         node.ID,
		Identifier: node.Identifier,
		Title:      node.Title,
		URL:        node.URL,
		StateName:  node.State.Name,
		StateType:  node.State.Type,
		TeamID:     node.Team.ID,
		TeamKey:    node.Team.Key,
		TeamName:   node.Team.Name,
	}, nil
}
