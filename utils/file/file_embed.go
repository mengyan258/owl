package file

import (
	"embed"
	"io"
	"os"
	"path/filepath"
)

// ExtractDir extracts a directory from the embedded filesystem.
func ExtractDir(embedFiles embed.FS, source string, destination string) error {
	entries, err := embedFiles.ReadDir(source)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := NormalizedPath(filepath.Join(source, entry.Name()))
		destinationPath := NormalizedPath(filepath.Join(destination, entry.Name()))

		if entry.IsDir() {
			if err := os.MkdirAll(destinationPath, 0755); err != nil {
				return err
			}
			if err := ExtractDir(embedFiles, sourcePath, destinationPath); err != nil {
				return err
			}
		} else {
			sourceFile, err := embedFiles.Open(sourcePath)
			if err != nil {
				return err
			}

			destinationFile, err := os.Create(destinationPath)
			if err != nil {
				return err
			}

			if _, err := io.Copy(destinationFile, sourceFile); err != nil {
				return err
			}
			_ = destinationFile.Close()
			_ = sourceFile.Close()
		}
	}

	return nil
}
