package processor

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func SplitFileBinary(inputFile string, chunkSize int64) ([]string, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	if fileInfo.Size() <= chunkSize {
		return []string{inputFile}, nil
	}

	var parts []string
	partNum := 1

	for {
		// Creates the part name: files.part1.zip
		ext := filepath.Ext(inputFile)
		base := strings.TrimSuffix(filepath.Base(inputFile), ext)
		partName := filepath.Join(filepath.Dir(inputFile), fmt.Sprintf("%s.part%d%s", base, partNum, ext))

		// Creates the file part
		partFile, err := os.Create(partName)
		if err != nil {
			return nil, fmt.Errorf("erro ao criar parte %s: %w", partName, err)
		}

		written, err := io.Copy(partFile, io.LimitReader(file, chunkSize))
		partFile.Close()

		if err != nil {
			return nil, fmt.Errorf("erro ao copiar dados para %s: %w", partName, err)
		}

		// if nothing is written, we finished in the last interation
		if written == 0 {
			os.Remove(partName)
			break
		}

		parts = append(parts, partName)
		partNum++

		if written < chunkSize {
			break
		}
	}

	return parts, nil
}
