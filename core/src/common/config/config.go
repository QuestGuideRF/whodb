package config
import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"github.com/clidey/whodb/core/src/common/datadir"
)
const configFileName = "config.json"
const (
	SectionCLI     = "cli"
	SectionAWS     = "aws"
	SectionAzure   = "azure"
	SectionGCP     = "gcp"
	SectionDesktop = "desktop"
)
var (
	configPath     string
	configPathOnce sync.Once
	configMu       sync.RWMutex
)
type RawConfig map[string]json.RawMessage
func getConfigPath(opts datadir.Options) (string, error) {
	var err error
	configPathOnce.Do(func() {
		var dir string
		dir, err = datadir.Get(opts)
		if err != nil {
			return
		}
		configPath = filepath.Join(dir, configFileName)
	})
	return configPath, err
}
func GetPath(opts datadir.Options) (string, error) {
	return getConfigPath(opts)
}
func ReadSection(section string, value any, opts datadir.Options) error {
	configMu.RLock()
	defer configMu.RUnlock()
	path, err := getConfigPath(opts)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var raw RawConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	sectionData, exists := raw[section]
	if !exists {
		return nil
	}
	return json.Unmarshal(sectionData, value)
}
func WriteSection(section string, value any, opts datadir.Options) error {
	configMu.Lock()
	defer configMu.Unlock()
	path, err := getConfigPath(opts)
	if err != nil {
		return err
	}
	var raw RawConfig
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		raw = make(RawConfig)
	} else {
		if err := json.Unmarshal(data, &raw); err != nil {
			raw = make(RawConfig)
		}
	}
	sectionData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	raw[section] = sectionData
	newData, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	tempFile, err := os.CreateTemp(dir, ".config-*.tmp")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	if _, err := tempFile.Write(newData); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return err
	}
	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath)
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return err
	}
	return nil
}
func ResetConfigPath() {
	configPathOnce = sync.Once{}
	configPath = ""
}