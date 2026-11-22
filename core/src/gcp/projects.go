package gcp
import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"github.com/clidey/whodb/core/src/log"
)
type LocalProject struct {
	ProjectID string
	Name      string
	Source    string
	IsDefault bool
}
func DiscoverLocalProjects() ([]LocalProject, error) {
	projects := make(map[string]*LocalProject)
	for _, envVar := range []string{"GOOGLE_CLOUD_PROJECT", "GCLOUD_PROJECT", "CLOUDSDK_CORE_PROJECT"} {
		if projectID := os.Getenv(envVar); projectID != "" {
			if _, exists := projects[projectID]; !exists {
				projects[projectID] = &LocalProject{
					ProjectID: projectID,
					Name:      projectID,
					Source:    "environment",
					IsDefault: true,
				}
			}
		}
	}
	if saPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); saPath != "" {
		if project := parseServiceAccountProject(saPath); project != nil {
			if _, exists := projects[project.ProjectID]; !exists {
				projects[project.ProjectID] = project
			}
		}
	}
	gcloudProjects := discoverGcloudProjects()
	for _, p := range gcloudProjects {
		if _, exists := projects[p.ProjectID]; !exists {
			projects[p.ProjectID] = &p
		}
	}
	result := make([]LocalProject, 0, len(projects))
	for _, project := range projects {
		result = append(result, *project)
	}
	return result, nil
}
func parseServiceAccountProject(path string) *LocalProject {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Debugf("GCP projects: failed to read service account file %s: %v", path, err)
		return nil
	}
	var sa struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(data, &sa); err != nil {
		log.Debugf("GCP projects: failed to parse service account file %s: %v", path, err)
		return nil
	}
	if sa.ProjectID == "" {
		return nil
	}
	return &LocalProject{
		ProjectID: sa.ProjectID,
		Name:      sa.ProjectID,
		Source:    "service-account",
		IsDefault: false,
	}
}
func discoverGcloudProjects() []LocalProject {
	configDir := getGcloudConfigDir()
	if configDir == "" {
		return nil
	}
	var projects []LocalProject
	propertiesPath := filepath.Join(configDir, "properties")
	if projectID := parseGcloudProperties(propertiesPath); projectID != "" {
		projects = append(projects, LocalProject{
			ProjectID: projectID,
			Name:      projectID,
			Source:    "gcloud-config",
			IsDefault: true,
		})
	}
	configurationsDir := filepath.Join(configDir, "configurations")
	entries, err := os.ReadDir(configurationsDir)
	if err != nil {
		return projects
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "config_") {
			continue
		}
		configPath := filepath.Join(configurationsDir, entry.Name())
		if projectID := parseGcloudProperties(configPath); projectID != "" {
			name := strings.TrimPrefix(entry.Name(), "config_")
			projects = append(projects, LocalProject{
				ProjectID: projectID,
				Name:      name,
				Source:    "gcloud-config",
				IsDefault: false,
			})
		}
	}
	return projects
}
func parseGcloudProperties(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	inCoreSection := false
	projectRegex := regexp.MustCompile(`^\s*project\s*=\s*(.+?)\s*$`)
	sectionRegex := regexp.MustCompile(`^\s*\[([^\]]+)\]\s*$`)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}
		if matches := sectionRegex.FindStringSubmatch(line); matches != nil {
			inCoreSection = matches[1] == "core"
			continue
		}
		if inCoreSection {
			if matches := projectRegex.FindStringSubmatch(line); matches != nil {
				return matches[1]
			}
		}
	}
	return ""
}
func getGcloudConfigDir() string {
	if dir := os.Getenv("CLOUDSDK_CONFIG"); dir != "" {
		return dir
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Warnf("Unable to determine home directory for gcloud config: %v", err)
		return ""
	}
	return filepath.Join(homeDir, ".config", "gcloud")
}