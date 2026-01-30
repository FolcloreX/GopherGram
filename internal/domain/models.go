package domain

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	MaxFileSize  int64 = 2 * 1024 * 1024 * 1024 // MaxFileSize 2GB
	HashTagVideo       = "F"                    // Ex: #F001
	HashTagDoc         = "Doc"                  // Ex: #Doc001
)

// Represents the entire struct of the Course being processed
type Course struct {
	RootPath string
	Modules  []*Module
	Assets   []string // Non videos
}

// Represente each folder of the Course being processed
type Module struct {
	Name     string
	Videos   []*Video
	Sequence int
}

type Video struct {
	FilePath string
	FileName string
	Title    string
	Module   string
	Duration int // Not being used currently saved for later
	Size     int64
	ID       string
	Sequence int
}

func (v *Video) FormatCaption() string {
	// Remove video extension to keep it clean
	cleanTitle := strings.TrimSuffix(v.Title, filepath.Ext(v.Title))

	return fmt.Sprintf("#%s %d - %s\n%s",
		v.ID,
		v.Sequence,
		cleanTitle,
		v.Module,
	)
}

func IsVideo(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp4", ".mkv", ".avi", ".mov", ".webm":
		return true
	default:
		return false
	}
}
