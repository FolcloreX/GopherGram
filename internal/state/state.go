package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

type StateData struct {
	TargetChatID int64           `json:"target_chat_id"`
	Processed    map[string]bool `json:"processed"`
}

type ProgressManager struct {
	FilePath string
	Data     *StateData
	mu       sync.Mutex
}

func Load(filename string) (*ProgressManager, error) {
	pm := &ProgressManager{
		FilePath: filename,
		Data: &StateData{
			TargetChatID: 0,
			Processed:    make(map[string]bool),
		},
	}

	if _, err := os.Stat(filename); err == nil {
		content, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(content, pm.Data); err != nil {
			return pm, nil
		}

		if pm.Data.Processed == nil {
			pm.Data.Processed = make(map[string]bool)
		}
	}

	return pm, nil
}

func LoadProgressContent(courseName string) (*ProgressManager, error) {
	// Slugify the file name
	re := regexp.MustCompile(`[^a-zA-Z0-9.-]+`)
	safeName := re.ReplaceAllString(courseName, "_")

	filename := fmt.Sprintf("progress_%s.json", safeName)
	path := filepath.Join("session", filename)

	//
	if err := os.MkdirAll("session", 0700); err != nil {
		return nil, fmt.Errorf("erro pasta session: %w", err)
	}

	return Load(path)
}

func (pm *ProgressManager) IsDone(filePath string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.Data.Processed[filePath]
}

func (pm *ProgressManager) MarkAsDone(filePath string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.Data.Processed[filePath] = true
	return pm.saveToDisk()
}

func (pm *ProgressManager) GetChatID() int64 {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.Data.TargetChatID
}

func (pm *ProgressManager) SetChatID(id int64) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.Data.TargetChatID = id
	return pm.saveToDisk()
}

func (pm *ProgressManager) saveToDisk() error {
	bytes, err := json.MarshalIndent(pm.Data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(pm.FilePath, bytes, 0644)
}
