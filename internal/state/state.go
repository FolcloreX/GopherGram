package state

import (
	"encoding/json"
	"os"
	"sync"
)

type ProgressManager struct {
	FilePath string
	Data     map[string]bool
	mu       sync.Mutex
}

func Load(filename string) (*ProgressManager, error) {
	pm := &ProgressManager{
		FilePath: filename,
		Data:     make(map[string]bool),
	}

	if _, err := os.Stat(filename); err == nil {
		content, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(content, &pm.Data); err != nil {
			// If json is corrupted start a new one
			return pm, nil
		}
	}

	return pm, nil
}

func (pm *ProgressManager) IsDone(filePath string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.Data[filePath]
}

func (pm *ProgressManager) MarkAsDone(filePath string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.Data[filePath] = true

	bytes, err := json.MarshalIndent(pm.Data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pm.FilePath, bytes, 0644)
}
