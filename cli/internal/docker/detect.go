package docker
import (
	"bufio"
	"bytes"
	"encoding/json"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"github.com/clidey/whodb/cli/internal/sourcetypes"
)
type DetectedContainer struct {
	Name        string
	Type        string
	Port        int
	Image       string
	ContainerID string
}
var imagePatterns = []struct {
	pattern *regexp.Regexp
	dbType  string
}{
	{regexp.MustCompile(`(?i)^postgres`), "Postgres"},
	{regexp.MustCompile(`(?i)^mysql`), "MySQL"},
	{regexp.MustCompile(`(?i)^mariadb`), "MariaDB"},
	{regexp.MustCompile(`(?i)^mongo`), "MongoDB"},
	{regexp.MustCompile(`(?i)^redis`), "Redis"},
	{regexp.MustCompile(`(?i)^clickhouse`), "ClickHouse"},
	{regexp.MustCompile(`(?i)^elasticsearch`), "ElasticSearch"},
	{regexp.MustCompile(`(?i)^docker\.elastic\.co/elasticsearch`), "ElasticSearch"},
	{regexp.MustCompile(`(?i)^duckdb`), "DuckDB"},
}
type dockerPSEntry struct {
	ID    string `json:"ID"`
	Image string `json:"Image"`
	Names string `json:"Names"`
	Ports string `json:"Ports"`
	State string `json:"State"`
}
func DetectContainers() []DetectedContainer {
	return detectContainersWithRunner(defaultRunner{})
}
type commandRunner interface {
	lookPath(file string) (string, error)
	run(name string, args ...string) ([]byte, error)
}
type defaultRunner struct{}
func (defaultRunner) lookPath(file string) (string, error) {
	return exec.LookPath(file)
}
func (defaultRunner) run(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}
func detectContainersWithRunner(runner commandRunner) []DetectedContainer {
	dockerPath, err := runner.lookPath("docker")
	if err != nil {
		return nil
	}
	out, err := runner.run(dockerPath, "ps", "--format", "{{json .}}")
	if err != nil {
		return nil
	}
	return parseDockerOutput(out)
}
func parseDockerOutput(data []byte) []DetectedContainer {
	var containers []DetectedContainer
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry dockerPSEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry.State != "running" {
			continue
		}
		dbType := matchImageType(entry.Image)
		if dbType == "" {
			continue
		}
		port := extractHostPort(entry.Ports, dbType)
		if port == 0 {
			continue
		}
		name := entry.Names
		if name == "" {
			name = entry.ID[:12]
		}
		containers = append(containers, DetectedContainer{
			Name:        name,
			Type:        dbType,
			Port:        port,
			Image:       entry.Image,
			ContainerID: entry.ID,
		})
	}
	return containers
}
func matchImageType(image string) string {
	parts := strings.Split(image, "/")
	for _, p := range imagePatterns {
		if p.pattern.MatchString(image) {
			return p.dbType
		}
	}
	last := parts[len(parts)-1]
	if idx := strings.Index(last, ":"); idx != -1 {
		last = last[:idx]
	}
	for _, p := range imagePatterns {
		if p.pattern.MatchString(last) {
			return p.dbType
		}
	}
	return ""
}
func defaultContainerPort(dbType string) int {
	port, ok := sourcetypes.DefaultPort(dbType)
	if !ok {
		return 0
	}
	return port
}
func extractHostPort(ports string, dbType string) int {
	containerPort := defaultContainerPort(dbType)
	if containerPort == 0 {
		return 0
	}
	containerPortStr := strconv.Itoa(containerPort)
	for _, mapping := range strings.Split(ports, ", ") {
		mapping = strings.TrimSpace(mapping)
		arrowIdx := strings.Index(mapping, "->")
		if arrowIdx == -1 {
			continue
		}
		right := mapping[arrowIdx+2:]
		if slashIdx := strings.Index(right, "/"); slashIdx != -1 {
			right = right[:slashIdx]
		}
		if right != containerPortStr {
			continue
		}
		left := mapping[:arrowIdx]
		colonIdx := strings.LastIndex(left, ":")
		if colonIdx == -1 {
			continue
		}
		hostPortStr := left[colonIdx+1:]
		hostPort, err := strconv.Atoi(hostPortStr)
		if err != nil {
			continue
		}
		return hostPort
	}
	return 0
}