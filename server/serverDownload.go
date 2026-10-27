package main

import(
	"encoding/binary"
	"io"
	"os"
	"archive/zip"
	"path/filepath"
)

func handleDownload(conn io.ReadWriter, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	totalSize := info.Size()

	var offset int64
	if err := binary.Read(conn, binary.BigEndian, &offset); err != nil {
		return err
	}

	if offset > totalSize {
		offset = totalSize
	}

	remainingSize := totalSize - offset

	if err := binary.Write(conn, binary.BigEndian, remainingSize); err != nil {
		return err
	}

	if remainingSize <= 0 {
		return nil
	}

	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	_, err = io.Copy(conn, file)
	return err
}
func handleDownloadDir(conn io.ReadWriter, dirPath string) error {
	tmpFile, err := os.CreateTemp("", "*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	zipWriter := zip.NewWriter(tmpFile)
	
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		zipWriter.Close()
		return err
	}
	
	if err := zipWriter.Close(); err != nil {
		return err
	}

	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return err
	}
	stat, err := tmpFile.Stat()
	if err != nil {
		return err
	}

	if err := binary.Write(conn, binary.BigEndian, stat.Size()); err != nil {
		return err
	}

	_, err = io.Copy(conn, tmpFile)
	return err
}