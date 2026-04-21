package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Draft   bool      `json:"draft"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

func (c Config) latestRelease(ctx context.Context) (*Release, error) {
	u := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", c.Owner, c.Repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api: %s", resp.Status)
	}

	var r ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	assets := make([]Asset, 0, len(r.Assets))
	for _, a := range r.Assets {
		assets = append(assets, Asset(a))
	}
	return &Release{TagName: r.TagName, Assets: assets}, nil
}
