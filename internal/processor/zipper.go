package processor

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Zipper struct {
	RootDir string
}

func (z *Zipper) ZipFiles(sourceFiles []string, destZip string) error {
	if len(sourceFiles) == 0 {
		return nil
	}

	newZipFile, err := os.Create(destZip)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo zip: %w", err)
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, filePath := range sourceFiles {
		if err := z.addFileToZip(zipWriter, filePath); err != nil {
			return err
		}
	}

	return nil
}

func (z *Zipper) addFileToZip(zipWriter *zip.Writer, filePath string) error {
	fileToZip, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo %s: %w", filePath, err)
	}
	defer fileToZip.Close()

	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Define the relative path inside de ZIP (Keep the file tree)
	// Ex: /home/user/course/Module1/pdf -> Modulo1/pdf
	relPath, err := filepath.Rel(z.RootDir, filePath)
	if err != nil {
		relPath = filepath.Base(filePath)
	}

	header.Name = filepath.ToSlash(relPath)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// Copy using stream, no file handled in the RAM
	_, err = io.Copy(writer, fileToZip)
	return err
}
