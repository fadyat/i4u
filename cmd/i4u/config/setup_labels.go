package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func SaveLabelsToYaml(filePath string, labels map[string]string) error {
	file, err := os.Create(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	defer func() {
		if e := file.Close(); e != nil {
			fmt.Printf("failed to close file: %v", e)
		}
	}()

	unitedLabels := map[string]any{"labels": labels}
	return yaml.NewEncoder(file).Encode(unitedLabels)
}
