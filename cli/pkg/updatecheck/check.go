package updatecheck
import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/clidey/whodb/cli/pkg/identity"
	goversion "github.com/hashicorp/go-version"
)
const (
	cacheTTL    = 24 * time.Hour
	httpTimeout = 1500 * time.Millisecond
)
type Result struct {
	LatestVersion   string
	UpdateAvailable bool
}
type cacheFile struct {
	LastCheck     time.Time `json:"lastCheck"`
	LatestVersion string    `json:"latestVersion"`
}
type githubRelease struct {
	TagName string `json:"tag_name"`
}
func Check(currentVersion string) *Result {
	if currentVersion == "" || currentVersion == "dev" {
		return nil
	}
	if os.Getenv("WHODB_DISABLE_UPDATE_CHECK") == "true" {
		return nil
	}
	cacheDir := getCacheDir()
	cachePath := filepath.Join(cacheDir, "update-check.json")
	var cached *cacheFile
	if currentCache, err := readCache(cachePath); err == nil {
		cached = currentCache
		if time.Since(cached.LastCheck) < cacheTTL {
			return compareVersions(currentVersion, cached.LatestVersion)
		}
	}
	latestVersion := currentVersion
	if cached != nil && cached.LatestVersion != "" {
		latestVersion = cached.LatestVersion
	}
	if release, err := fetchLatestRelease(); err == nil {
		latestVersion = release.TagName
	}
	_ = os.MkdirAll(cacheDir, 0700)
	_ = writeCache(cachePath, &cacheFile{
		LastCheck:     time.Now(),
		LatestVersion: latestVersion,
	})
	return compareVersions(currentVersion, latestVersion)
}
func compareVersions(currentVersion, latestTag string) *Result {
	latestClean := strings.TrimPrefix(latestTag, "v")
	currentClean := strings.TrimPrefix(currentVersion, "v")
	current, err := goversion.NewVersion(currentClean)
	if err != nil {
		return nil
	}
	latest, err := goversion.NewVersion(latestClean)
	if err != nil {
		return nil
	}
	if latest.GreaterThan(current) {
		return &Result{
			LatestVersion:   latestTag,
			UpdateAvailable: true,
		}
	}
	return nil
}
func getCacheDir() string {
	cacheDir, err := identity.HomePath()
	if err != nil {
		baseName := strings.TrimPrefix(identity.Current().HomeDirName, ".")
		if baseName == "" {
			return os.TempDir()
		}
		return filepath.Join(os.TempDir(), baseName)
	}
	return cacheDir
}
func readCache(path string) (*cacheFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c cacheFile
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
func writeCache(path string, c *cacheFile) error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
func fetchLatestRelease() (*githubRelease, error) {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(identity.Current().UpdateCheckAPIURL)
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