package main

import(
	"encoding/binary"
	"io"
	"os"
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

	if err := binary.Write(conn, binary.BigEndian, info.Size()); err != nil {
		return err
	}

	_, err = io.Copy(conn, file)
	return err
}