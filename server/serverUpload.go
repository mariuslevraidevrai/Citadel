package main

import(
	"encoding/binary"
	"io"
	"os"
)

func handleUpload(conn io.Reader, destPath string) error {
	var size int64
	if err := binary.Read(conn, binary.BigEndian, &size); err != nil {
		return err
	}

	file, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.CopyN(file, conn, size)
	return err
}