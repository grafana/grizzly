package grizzly

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"

	"github.com/minio/selfupdate"
	"golang.org/x/mod/semver"
)

var ErrCurrentVersionIsLatest = fmt.Errorf("current version is the latest")
var ErrNextVersionIsMajorBump = fmt.Errorf("next version is a major bump")
var ErrInvalidSemver = fmt.Errorf("invalid semver version")

type ghRelease struct {
	Draft      bool             `json:"draft"`
	PreRelease bool             `json:"prerelease"`
	TagName    string           `json:"tag_name"`
	Assets     []ghReleaseAsset `json:"assets"`
}

func (r ghRelease) assetFor(goos string, goarch string) (string, bool) {
	expectedAssetNameSuffix := fmt.Sprintf("%s-%s", goos, goarch)

	for _, asset := range r.Assets {
		if !strings.HasSuffix(asset.Name, expectedAssetNameSuffix) {
			continue
		}

		return asset.DownloadURL, true
	}

	return "", false
}

type ghReleaseAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

type SelfUpdater struct {
	http *http.Client
}

func NewSelfUpdater(client *http.Client) *SelfUpdater {
	return &SelfUpdater{
		http: client,
	}
}

func (updater *SelfUpdater) UpdateSelf(ctx context.Context, currentVersion string) (string, error) {
	if !semver.IsValid(currentVersion) {
		return "", fmt.Errorf("invalid current version '%s': %w", currentVersion, ErrInvalidSemver)
	}

	latestRelease, err := updater.latestStableRelease(ctx)
	if err != nil {
		return "", err
	}

	if !semver.IsValid(latestRelease.TagName) {
		return "", fmt.Errorf("invalid latest version '%s': %w", latestRelease.TagName, ErrInvalidSemver)
	}

	// latest is <= currentVersion
	if semver.Compare(latestRelease.TagName, currentVersion) <= 0 {
		return currentVersion, ErrCurrentVersionIsLatest
	}

	if semver.Major(latestRelease.TagName) != semver.Major(currentVersion) {
		return latestRelease.TagName, ErrNextVersionIsMajorBump
	}

	assetURL, found := latestRelease.assetFor(runtime.GOOS, runtime.GOARCH)
	if !found {
		return "", fmt.Errorf("could not find binary for version=%s, GOOS=%s, GOARCH=%s", latestRelease.TagName, runtime.GOOS, runtime.GOARCH)
	}

	if err := updater.doUpdate(ctx, assetURL); err != nil {
		return "", err
	}

	return latestRelease.TagName, nil
}

func (updater *SelfUpdater) doUpdate(ctx context.Context, assetURL string) error {
	response, err := updater.httpGet(ctx, assetURL)
	defer func() {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
	}()
	if err != nil {
		return err
	}

	return selfupdate.Apply(response.Body, selfupdate.Options{})
}

func (updater *SelfUpdater) latestStableRelease(ctx context.Context) (*ghRelease, error) {
	response, err := updater.httpGet(ctx, "https://api.github.com/repos/grafana/grizzly/releases/latest")
	defer func() {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
	}()
	if err != nil {
		return nil, err
	}

	release := &ghRelease{}
	if err := json.NewDecoder(response.Body).Decode(release); err != nil {
		return nil, err
	}

	return release, nil
}

func (updater *SelfUpdater) httpGet(ctx context.Context, url string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	response, err := updater.http.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, updater.httpError(response)
	}

	return response, nil
}

func (updater *SelfUpdater) httpError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return fmt.Errorf("unexpected HTTP response (HTTP status %d): %s ", resp.StatusCode, body)
}
