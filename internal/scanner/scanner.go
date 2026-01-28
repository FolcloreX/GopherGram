package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	domain "github.com/FolcloreX/GopherGram/internal/domains"
)

type Scanner struct {
	RootPath string
}

func New(rootPath string) *Scanner {
	return &Scanner{RootPath: rootPath}
}

// Scan go through the folder and construct the file tree
func (s *Scanner) Scan() (*domain.Course, error) {
	course := &domain.Course{
		RootPath: s.RootPath,
		Modules:  []*domain.Module{},
		Assets:   []string{},
	}

	entries, err := os.ReadDir(s.RootPath)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler diretório raiz: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return naturalLess(entries[i].Name(), entries[j].Name())
	})

	globalVideoSequence := 1

	for _, entry := range entries {
		fullPath := filepath.Join(s.RootPath, entry.Name())

		// TODO for now, everything that's not a folder in the root will be considered an asset
		if !entry.IsDir() {
			course.Assets = append(course.Assets, fullPath)
			continue
		}

		// If it's a diretory -> New Module
		module := &domain.Module{
			Name:   entry.Name(),
			Videos: []*domain.Video{},
		}

		err := filepath.WalkDir(fullPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			if domain.IsVideo(d.Name()) {
				vid := &domain.Video{
					FilePath: path,
					FileName: d.Name(),
					Title:    d.Name(), // Formated latter
					Module:   module.Name,
					ID:       fmt.Sprintf("%s%03d", domain.HashTagVideo, globalVideoSequence),
					Sequence: globalVideoSequence,
				}

				info, _ := d.Info()
				if info != nil {
					vid.Size = info.Size()
				}

				module.Videos = append(module.Videos, vid)
				globalVideoSequence++
			} else {
				course.Assets = append(course.Assets, path)
			}
			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("erro ao varrer módulo %s: %w", module.Name, err)
		}

		// Order the videos inside the module
		sort.Slice(module.Videos, func(i, j int) bool {
			return naturalLess(module.Videos[i].FileName, module.Videos[j].FileName)
		})

		// Only add the module if there's content
		if len(module.Videos) > 0 {
			course.Modules = append(course.Modules, module)
		}
	}

	return course, nil
}
