package main

import(
	"os"
	"encoding/binary"
	"net"
	"io"
)

func clientUpload(conn net.Conn, localPath string) error {
	file, err := os.Open(localPath)
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