package common
import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"github.com/clidey/whodb/core/src/engine"
)
func GetRecordValueOrDefault(records []engine.Record, key string, defaultValue string) string {
	for _, record := range records {
		if record.Key == key && len(record.Value) > 0 {
			return record.Value
		}
	}
	return defaultValue
}
func IsRunningInsideDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return !os.IsNotExist(err)
}
func IsRunningInsideWSL2() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	version := strings.ToLower(string(data))
	return strings.Contains(version, "microsoft") || strings.Contains(version, "wsl")
}
func GetWSL2WindowsHost() string {
	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 || fields[1] != "00000000" {
			continue
		}
		gw, err := strconv.ParseUint(fields[2], 16, 32)
		if err != nil {
			return ""
		}
		return fmt.Sprintf("%d.%d.%d.%d", gw&0xFF, (gw>>8)&0xFF, (gw>>16)&0xFF, (gw>>24)&0xFF)
	}
	return ""
}
func replaceHost(hostWithPort, oldHost, newHost string) string {
	return strings.Replace(hostWithPort, oldHost, newHost, 1)
}
func ResolveLocalURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	hostname := parsed.Hostname()
	if hostname != "localhost" && hostname != "127.0.0.1" {
		return rawURL
	}
	if IsRunningInsideDocker() {
		parsed.Host = replaceHost(parsed.Host, hostname, "host.docker.internal")
		return parsed.String()
	}
	if IsRunningInsideWSL2() {
		if wslHost := GetWSL2WindowsHost(); wslHost != "" {
			parsed.Host = replaceHost(parsed.Host, hostname, wslHost)
			return parsed.String()
		}
	}
	return rawURL
}
func FilterList[T any](items []T, by func(input T) bool) []T {
	filteredItems := []T{}
	for _, item := range items {
		if by(item) {
			filteredItems = append(filteredItems, item)
		}
	}
	return filteredItems
}
func StrPtrToBool(s *string) bool {
	if s == nil {
		return false
	}
	value := strings.ToLower(*s)
	return value == "true"
}