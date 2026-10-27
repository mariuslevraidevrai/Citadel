package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"crypto/tls"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"golang.org/x/term"
)

const Banner = `
   █████████   ███   █████                  █████          ████ 
  ███▒▒▒▒▒███ ▒▒▒   ▒▒███                  ▒▒███          ▒▒███ 
 ███     ▒▒▒  ████  ███████    ██████    ███████   ██████  ▒███ 
▒███         ▒▒███ ▒▒▒███▒    ▒▒▒▒▒███  ███▒▒███  ███▒▒███ ▒███ 
▒███          ▒███   ▒███      ███████ ▒███ ▒███ ▒███████  ▒███ 
▒▒███     ███ ▒███   ▒███ ███ ███▒▒███ ▒███ ▒███ ▒███▒▒▒   ▒███ 
 ▒▒█████████  █████  ▒▒█████ ▒▒████████▒▒████████▒▒██████  █████
  ▒▒▒▒▒▒▒▒▒  ▒▒▒▒▒    ▒▒▒▒▒   ▒▒▒▒▒▒▒▒  ▒▒▒▒▒▒▒▒  ▒▒▒▒▒▒  ▒▒▒▒▒ 
                                                                
`
const Version = "0.0.1"

func main() {
	ip := flag.String("ip", "127.0.0.1", "Server IP address")
	port := flag.Int("port", 9000, "Server port")
	version := flag.Bool("version", false, "Citadel version")

	flag.Parse()
	if *version{
		fmt.Print(Banner)
		fmt.Printf("Citadel %s\n\n",Version)
		fmt.Println("A very fast and secure file server")
		fmt.Println("For more info visit github.com/mariuslevraidevrai/citadel")
		return
	}

	if *port < 1 || *port > 65535 {
		log.Fatal(color.YellowString("Port must be between 1 and 65535"))
	}
	address := net.JoinHostPort(*ip, strconv.Itoa(*port))

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", address, tlsConfig)
	if err != nil {
		log.Fatal(color.RedString("Failed to connect to Citadel server: ", err))
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

    fmt.Print("Enter password: ")

    passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
    if err != nil {
        panic(err)
    }

    password := strings.TrimSpace(string(passwordBytes))

    _ = password

	fmt.Fprintf(writer, "AUTH %s\n", password)
	writer.Flush()

	responseAuth, err := reader.ReadString('\n')
	if err != nil || !strings.Contains(responseAuth, "AUTH_OK") {
		fmt.Println(color.RedString("error: authentification failed"))
		return
	}
	fmt.Println(color.GreenString("\nSuccessful authentification !"))

	promptPath := color.BlueString("/")

	completer := readline.NewPrefixCompleter(
		readline.PcItem("help"),
		readline.PcItem("ping"),
		readline.PcItem("pwd"),
		readline.PcItem("ls"),
		readline.PcItem("cd"),
		readline.PcItem("cat"),
		readline.PcItem("tree"),
		readline.PcItem("info"),
		readline.PcItem("rename"),
		readline.PcItem("mkdir"),
		readline.PcItem("touch"),
		readline.PcItem("remove"),
		readline.PcItem("copy"),
		readline.PcItem("move"),
		readline.PcItem("upload"),
		readline.PcItem("download"),
		readline.PcItem("status"),
		readline.PcItem("quit"),
	)

	terminal, err := readline.NewEx(&readline.Config{
		Prompt:          "\n" + promptPath + " > ",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		log.Fatal(color.RedString("Failed to initialize terminal: ", err))
	}
	defer terminal.Close()

	fmt.Printf(
		"Connected to %s:%d. Type 'help' to see available commands.\n\n",
		*ip,
		*port,
	)

	for {
		terminal.SetPrompt("(" + promptPath + ") > ")

		command, err := terminal.Readline()
		if err != nil {
			if errors.Is(err, readline.ErrInterrupt) {
				continue
			}
			if errors.Is(err, io.EOF) {
				fmt.Println()
				return
			}

			log.Println(color.RedString("Failed to read terminal input:", err))
			return
		}

		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}

		fields := strings.Fields(command)
		action := strings.ToLower(fields[0])

		if action == "ping" {
			start := time.Now()
			if _, err := fmt.Fprintln(writer, command); err != nil {
				log.Println(color.RedString("Failed to send command:", err))
				return
			}
			if err := writer.Flush(); err != nil {
				log.Println(color.RedString("Failed to flush command:", err))
				return
			}

			response, err := readResponse(reader)
			latency := time.Since(start)
			if err != nil {
				log.Println(color.RedString("Failed to read server response:", err))
				return
			}
			if response != "" {
				fmt.Printf("%s %s\n", response, fmt.Sprintf("(Latency : %v)", latency))
			}
			continue
		}

		if action == "status" {
			start := time.Now()
			if _, err := fmt.Fprintln(writer, "ping"); err != nil {
				fmt.Println(color.RedString("error: server is offline"))
				continue
			}
			if err := writer.Flush(); err != nil {
				fmt.Println(color.RedString("error: failed to flush status check"))
				continue
			}

			_, err = readResponse(reader)
			latency := time.Since(start)

			onlineStatus := color.GreenString("Online")
			if err != nil {
				onlineStatus = color.RedString("Offline")
				latency = 0
			}

			fmt.Println(color.YellowString("Server status: "))
			fmt.Printf("  %s %s\n", color.GreenString("State :"), onlineStatus)
			fmt.Printf("  %s %s\n", color.GreenString("IP Address :"), *ip)
			fmt.Printf("  %s %d\n", color.GreenString("Port :"), *port)
			fmt.Printf("  %s %v\n", color.GreenString("Latency :"), latency)
			continue
		}

		if action == "upload" {
			if len(fields) < 2 || len(fields) > 3 {
				fmt.Println(color.YellowString("usage: upload <local_file> [remote_name]"))
				continue
			}
			localFile := fields[1]

			if _, err := os.Stat(localFile); err != nil {
				fmt.Println(color.RedString("error: local file does not exist"))
				continue
			}

			if _, err := fmt.Fprintln(writer, command); err != nil {
				log.Println(color.RedString("Failed to send command:", err))
				return
			}
			writer.Flush()

			sig, err := reader.ReadString('\n')
			if err != nil {
				log.Println(color.RedString("Failed to receive server signal:", err))
				return
			}

			sigTrim := strings.TrimSpace(sig)
			if !strings.HasPrefix(sigTrim, "READY") {
				response, _ := readResponseCustom(reader, sigTrim)
				fmt.Println(response)
				continue
			}

			if err := clientUpload(conn, localFile); err != nil {
				fmt.Println(color.RedString("error: upload failed: " + err.Error()))
			} else {
				fmt.Println(color.GreenString("File uploaded successfully"))
			}

			response, err := readResponse(reader)
			if err == nil && response != "" {
				fmt.Println(response)
			}
			continue
		}

		if action == "download" {
			if len(fields) < 2 || len(fields) > 3 {
				fmt.Println(color.YellowString("usage: download <remote_file> [local_name]"))
				continue
			}

			if _, err := fmt.Fprintln(writer, command); err != nil {
				log.Println(color.RedString("Failed to send command:", err))
				return
			}
			writer.Flush()

			sig, err := reader.ReadString('\n')
			if err != nil {
				log.Println(color.RedString("Failed to receive server signal:", err))
				return
			}

			sigTrim := strings.TrimSpace(sig)
			if !strings.HasPrefix(sigTrim, "READY") {
				response, _ := readResponseCustom(reader, sigTrim)
				fmt.Println(response)
				continue
			}

			remoteFile := fields[1]
			localName := filepath.Base(remoteFile)
			if len(fields) == 3 {
				localName = fields[2]
			}

			if err := clientDownload(conn, localName); err != nil {
				fmt.Println(color.RedString("error: download failed: " + err.Error()))
			} else {
				fmt.Println(color.GreenString("File downloaded successfully: " + localName))
			}
			continue
		}

		if _, err := fmt.Fprintln(writer, command); err != nil {
			log.Println(color.RedString("Failed to send command:", err))
			return
		}

		if err := writer.Flush(); err != nil {
			log.Println(color.RedString("Failed to flush command:", err))
			return
		}

		response, err := readResponse(reader)
		if err != nil {
			log.Println(color.RedString("Failed to read server response:", err))
			return
		}

		if strings.HasPrefix(response, "cwd: ") {
			promptPath = strings.TrimPrefix(response, "cwd: ")
		} else if response != "" {
			fmt.Println(response)
		}

		if strings.EqualFold(command, "quit") || strings.EqualFold(command, "bye") {
			return
		}
	}
}

func readResponse(reader *bufio.Reader) (string, error) {
	var lines []string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		line = strings.TrimRight(line, "\r\n")

		if line == "END" {
			return strings.Join(lines, "\n"), nil
		}

		lines = append(lines, line)
	}
}

func readResponseCustom(reader *bufio.Reader, firstLine string) (string, error) {
	lines := []string{firstLine}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return strings.Join(lines, "\n"), err
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "END" {
			break
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n"), nil
}
