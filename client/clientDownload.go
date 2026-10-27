package main

import (
	"encoding/binary"
	"io"
	"os"
)

func clientDownload(conn io.ReadWriter, localName string) error {
	var offset int64 = 0

	if info, err := os.Stat(localName); err == nil && !info.IsDir() {
		offset = info.Size()
	}

	if err := binary.Write(conn, binary.BigEndian, offset); err != nil {
		return err
	}

	var remainingSize int64
	if err := binary.Read(conn, binary.BigEndian, &remainingSize); err != nil {
		return err
	}

	if remainingSize == 0 {
		return nil
	}

	file, err := os.OpenFile(localName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.CopyN(file, conn, remainingSize)
	return err
}