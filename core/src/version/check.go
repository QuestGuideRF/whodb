package version
import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	goversion "github.com/hashicorp/go-version"
)
const (
	githubReleasesURL = "https://api.github.com/repos/clidey/whodb/releases/latest"
	cacheTTL          = 24 * time.Hour
	httpTimeout       = 5 * time.Second
)
type UpdateInfo struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	ReleaseURL      string
}
type cachedResult struct {
	info      UpdateInfo
	checkedAt time.Time
}
var (
	cache               *cachedResult
	cacheMu             sync.Mutex
	latestReleaseGetter = fetchLatestRelease
)
type githubRelease struct {
	TagName string `json:"tag_name"`
}
func CheckForUpdate(currentVersion string, disabled bool) UpdateInfo {
	noUpdate := UpdateInfo{
		CurrentVersion:  currentVersion,
		LatestVersion:   currentVersion,
		UpdateAvailable: false,
	}
	if disabled || currentVersion == "" || currentVersion == "development" {
		return noUpdate
	}
	cacheMu.Lock()
	if cache != nil && time.Since(cache.checkedAt) < cacheTTL {
		result := cache.info
		cacheMu.Unlock()
		return result
	}
	cacheMu.Unlock()
	result := noUpdate
	if release, err := latestReleaseGetter(); err == nil {
		latestTag := strings.TrimPrefix(release.TagName, "v")
		currentClean := strings.TrimPrefix(currentVersion, "v")
		current, currErr := goversion.NewVersion(currentClean)
		latest, latErr := goversion.NewVersion(latestTag)
		if currErr == nil && latErr == nil && latest.GreaterThan(current) {
			result = UpdateInfo{
				CurrentVersion:  currentVersion,
				LatestVersion:   release.TagName,
				UpdateAvailable: true,
				ReleaseURL:      fmt.Sprintf("https://github.com/clidey/whodb/releases/tag/%s", release.TagName),
			}
		}
	}
	cacheMu.Lock()
	cache = &cachedResult{info: result, checkedAt: time.Now()}
	cacheMu.Unlock()
	return result
}
func fetchLatestRelease() (*githubRelease, error) {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(githubReleasesURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected status")
	}
	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}