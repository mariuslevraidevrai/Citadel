
# Citadel

<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Linux-FCC624?logo=linux&logoColor=black" alt="Linux">
  <img src="https://img.shields.io/badge/GNU_Make-A42E2B?logo=gnu&logoColor=white" alt="GNU Make">
  <img src="https://img.shields.io/badge/TLS-Encrypted-blue" alt="TLS">
  <img src="https://img.shields.io/badge/License-GPL--3.0-blue" alt="GPL v3.0">
</p>

<p align="center">
  <strong>A fast, lightweight file server written in Go.</strong>
</p>

Citadel is a client-server file management system featuring TLS-encrypted communication, password authentication and remote file operations. It supports Linux and Windows across multiple architectures.

## Features

- TLS-encrypted communication
- Password-based authentication
- File uploads and downloads
- Remote file and directory management
- Interactive CLI with command autocomplete
- Server status and latency checks
- Cross-compilation for Linux and Windows

## Installation

### Requirements

- Go
- GNU Make
- Git

**Debian / Ubuntu**
```bash
sudo apt update
sudo apt install -y golang-go make git
```

**Fedora**
```bash
sudo dnf install -y golang make git
```

### Build from source

```bash
git clone https://github.com/mariuslevraidevrai/citadel.git
cd citadel
make setup
make build
```

Binaries are generated in `bin/`.

## Usage

### Start the server

```bash
./bin/citadel-server -ip 0.0.0.0 -port 9000 -password "YOUR_PASSWORD" -dir shared
```

If the dir flag is not set the server creates a `shared/` default directory for remotely accessible files and generates a self-signed TLS certificate.

### Connect with the client

```bash
./bin/citadel-client -ip 192.168.1.28 -port 9000
```

Replace the IP address with your server's address. Enter the password when prompted.

### Available commands

| Command | Description |
|---|---|
| `help` | Show help |
| `ls`, `tree` | Browse files |
| `cd`, `pwd` | Navigate directories |
| `cat`, `info` | Read files and inspect metadata |
| `search` | Search a file |
| `mkdir`, `touch` | Create directories and files |
| `rename`, `copy`, `move`, `remove` | Manage files |
| `upload`, `download` | Transfer files |
| `ping`, `status` | Check connectivity |
| `quit` | Disconnect |

## Cross-compilation

```bash
make build          # Current platform
make linux          # Linux AMD64, ARM64 and ARMv7
make windows        # Windows AMD64 and ARM64
make all            # All supported targets
make clean          # Remove binaries
make help           # Show Makefile targets
```

## Security

Citadel uses TLS and password authentication. For secure deployments:

- Use a strong, unique password.
- Restrict network access with a firewall.
- Protect the server's private key and shared files.
- Enable TLS certificate verification in the client. The current client disables certificate verification, leaving connections vulnerable to man-in-the-middle attacks.

## License

Citadel is licensed under the **GNU General Public License v3.0 (GPL-3.0)**.

See the [LICENSE](LICENSE) file for details.

---

<p align="center">
  <strong>Citadel — Fast. Lightweight. Open source.</strong>
</p>
