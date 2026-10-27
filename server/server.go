package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"crypto/tls"

	"github.com/fatih/color"
)

const sharedDir = "shared"
const maxLineSize = 64 * 1024
const maxFileSize = 64 * 1024

func main() {
	ip := flag.String("ip", "127.0.0.1", "IP address to listen on")
	port := flag.Int("port", 9000, "Port to listen on")
	password := flag.String("password", "citadel", "Password to auth on the server")
	
	flag.Parse()

	if *port < 1 || *port > 65535 {
		log.Fatal(color.YellowString("Port must be between 1 and 65535"))
	}

	if err := os.MkdirAll(sharedDir, 0750); err != nil {
		log.Fatal(color.RedString("Failed to create shared directory: ", err))
	}

	setPassword(*password)

	root, err := filepath.Abs(sharedDir)
	if err != nil {
		log.Fatal(color.RedString("Failed to resolve shared directory: ", err))
	}

	address := net.JoinHostPort(*ip, strconv.Itoa(*port))
	if err := generateSelfSignedCert(); err != nil {
		log.Fatalf("Failed to generate certificates : %v", err)
	}

	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatal(color.RedString("Failed to load certificates: ", err))
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	listener, err := tls.Listen("tcp", address, tlsConfig)
	if err != nil {
		log.Fatal(color.RedString("Failed to start server: ", err))
	}
	defer listener.Close()

	log.Printf(color.YellowString("Citadel server is listening on %s\n", address))
	log.Printf(color.BlueString("Shared directory: %s", root))

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(color.RedString("Failed to accept connection:", err))
			continue
		}

		go handleConnection(conn, root)
	}
}

func handleConnection(conn net.Conn, root string) {
	defer conn.Close()
	log.Printf(color.GreenString("Connection received from %s", conn.RemoteAddr()))

	scanner := bufio.NewScanner(conn)
	writer := bufio.NewWriter(conn)

	if !scanner.Scan() {
		return
	}
	authCmd := scanner.Text()
	fields := strings.Fields(authCmd)

	if len(fields) != 2 || strings.ToLower(fields[0]) != "auth" {
		fmt.Fprintln(writer, color.RedString("error: authentication required"))
		writer.Flush()
		return
	}

	valid, err := checkPassword(fields[1])
	if err != nil || !valid {
		fmt.Fprintln(writer, color.RedString("error: invalid password"))
		writer.Flush()
		log.Println(color.RedString("Authentification failed for %s", conn.RemoteAddr()))
		return
	}

	fmt.Fprintln(writer, "AUTH_OK")
	writer.Flush()
	log.Println(color.GreenString("Authentificated client : %s", conn.RemoteAddr()))

	cwd := "."
	for scanner.Scan() {
		command := scanner.Text()
		response, shouldClose, newCwd := handleCommand(conn, writer, command, &cwd, root)

		if newCwd != "" {
			cwd = newCwd
		}

		if response != "" {
			if _, err := fmt.Fprintln(writer, response); err != nil {
				log.Printf(color.RedString("Failed to send response: %v", err))
				return
			}
			if _, err := fmt.Fprintln(writer, "END"); err != nil {
				log.Printf(color.RedString("Failed to send response terminator: %v", err))
				return
			}
			if err := writer.Flush(); err != nil {
				log.Printf(color.RedString("Failed to flush response: %v", err))
				return
			}
		}

		if shouldClose {
			return
		}
	}
}

func handleCommand(conn net.Conn, writer *bufio.Writer, command string, cwd *string, root string) (response string, shouldClose bool, newCwd string) {
	fields := strings.Fields(strings.TrimSpace(command))
	if len(fields) == 0 {
		return color.RedString("error: empty command"), false, ""
	}

	action := strings.ToLower(fields[0])
	args := fields[1:]

	switch action {
	case "help":
		return help(args), false, ""

	case "ping":
		return color.GreenString("Ping from "+ conn.RemoteAddr().String()+" to "+conn.LocalAddr().String()), false, ""

	case "pwd":
		return displayPath(*cwd), false, ""

	case "ls":
			target := "."
			if len(args) == 1 {
				target = args[0]
			} else if len(args) > 1 {
				return color.YellowString("usage: ls [directory]"), false, ""
			}
			return listDirectory(root, *cwd, target), false, ""

	case "cd":
		if len(args) != 1 {
			return color.YellowString("usage: cd <directory>"), false, ""
		}
		path, err := resolveSharedPath(root, *cwd, args[0])
		if err != nil || !isDir(path) {
			return color.RedString("error: invalid directory"), false, ""
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return color.RedString("error: could not change directory"), false, ""
		}
		if rel == "" {
			rel = "."
		}
		return "cwd: " + cleanPath(rel), false, rel

	case "cat":
		if len(args) != 1 {
			return color.YellowString("usage: cat <file>"), false, ""
		}
		return readFile(root, *cwd, args[0], maxFileSize), false, ""

	case "tree":
		target := "."
		if len(args) == 1 {
			target = args[0]
		}
		return treeDirectory(root, *cwd, target), false, ""

	case "info":
		if len(args) != 1 {
			return color.YellowString("usage: info <name>"), false, ""
		}
		return fileInfo(root, *cwd, args[0]), false, ""

	case "rename":
		if len(args) != 2 {
			return color.YellowString("usage: rename <old> <new>"), false, ""
		}
		oldPath, err1 := resolveSharedPath(root, *cwd, args[0])
		newPath, err2 := resolveSharedPath(root, *cwd, args[1])
		if err1 != nil || err2 != nil {
			return color.RedString("error: invalid path"), false, ""
		}
		if err := os.Rename(oldPath, newPath); err != nil {
			return color.RedString("error: " + err.Error()), false, ""
		}
		return color.GreenString("Renamed successfully"), false, ""

	case "mkdir":
		if len(args) != 1 {
			return color.YellowString("usage: mkdir <name>"), false, ""
		}
		path, err := resolveSharedPath(root, *cwd, args[0])
		if err != nil {
			return color.RedString("error: invalid path"), false, ""
		}
		if err := os.Mkdir(path, 0750); err != nil {
			return color.RedString("error: " + err.Error()), false, ""
		}
		return color.GreenString("Directory created"), false, ""

	case "touch":
		if len(args) != 1 {
			return color.YellowString("usage: touch <name>"), false, ""
		}
		path, err := resolveSharedPath(root, *cwd, args[0])
		if err != nil {
			return color.RedString("error: invalid path"), false, ""
		}
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
		if err != nil {
			return color.RedString("error: " + err.Error()), false, ""
		}
		file.Close()
		return color.GreenString("File created"), false, ""

	case "remove":
		if len(args) != 1 {
			return color.YellowString("usage: remove <name>"), false, ""
		}
		path, err := resolveSharedPath(root, *cwd, args[0])
		if err != nil || path == root {
			return color.RedString("error: invalid path"), false, ""
		}
		if err := os.Remove(path); err != nil {
			return color.RedString("error: " + err.Error()), false, ""
		}
		return color.GreenString("Removed successfully"), false, ""

	case "copy", "move":
		if len(args) != 2 {
			return color.YellowString("usage: " + action + " <source> <destination>"), false, ""
		}
		src, err1 := resolveSharedPath(root, *cwd, args[0])
		dst, err2 := resolveSharedPath(root, *cwd, args[1])
		if err1 != nil || err2 != nil || src == root {
			return color.RedString("error: invalid path"), false, ""
		}
		if action == "copy" {
			if err := copyFile(src, dst); err != nil {
				return color.RedString("error: " + err.Error()), false, ""
			}
			return color.GreenString("Copied successfully"), false, ""
		}
		if err := os.Rename(src, dst); err != nil {
			return color.RedString("error: " + err.Error()), false, ""
		}
		return color.GreenString("Moved successfully"), false, ""

	case "download":
		if len(args) < 1 || len(args) > 2 {
			return color.YellowString("usage: download <remote_file> [local_name]"), false, ""
		}
		path, err := resolveSharedPath(root, *cwd, args[0])
		if err != nil {
			writer.WriteString("ERROR: invalid path\nEND\n")
			writer.Flush()
			return "", false, ""
		}
		if _, err := os.Stat(path); err != nil {
			writer.WriteString("ERROR: file does not exist\nEND\n")
			writer.Flush()
			return "", false, ""
		}
		writer.WriteString("READY\n")
		writer.Flush()
		if err := handleDownload(conn, path); err != nil {
			log.Printf("Download error: %v", err)
		}
		return "", false, ""

	case "upload":
		if len(args) < 1 || len(args) > 2 {
			return color.YellowString("usage: upload <local_file> [remote_name]"), false, ""
		}
		targetName := args[0]
		if len(args) == 2 {
			targetName = args[1]
		}
		path, err := resolveSharedPath(root, *cwd, targetName)
		if err != nil {
			writer.WriteString("ERROR: invalid path\nEND\n")
			writer.Flush()
			return "", false, ""
		}
		writer.WriteString("READY\n")
		writer.Flush()
		if err := handleUpload(conn, path); err != nil {
			log.Printf("Upload error: %v", err)
			return color.RedString("error: upload failed"), false, ""
		}
		return "", false, ""

	case "quit", "bye":
		return "\nGoodbye", true, ""

	default:
		return color.RedString("error: unknown command"), false, ""
	}
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}