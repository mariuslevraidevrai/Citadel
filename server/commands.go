package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
)

func help(args []string) string {
	commands := map[string]string{
		"help":     "help [command] - show help",
		"ping":     "ping - test connection latency",
		"pwd":      "pwd - show current directory",
		"cd":       "cd <directory> - change directory",
		"ls":       "ls - list directory contents",
		"cat":      "cat <file> - display a file",
		"tree":     "tree [directory] - display directory tree",
		"info":     "info <name> - display item details",
		"rename":   "rename <old> <new> - rename an item",
		"mkdir":    "mkdir <name> - create a directory",
		"touch":    "touch <name> - create an empty file",
		"remove":   "remove <name> - delete a file or empty directory",
		"copy":     "copy <source> <destination> - copy a file",
		"move":     "move <source> <destination> - move an item",
		"upload":   "upload <local_file> [remote_name] - upload a file to the server",
		"download": "download <remote_file> [local_name] - download a file from the server",
		"status":   "status - display server connection details",
		"quit":     "quit - disconnect",
	}

	if len(args) == 1 {
		if description, ok := commands[strings.ToLower(args[0])]; ok {
			return color.YellowString(description)
		}
		return color.RedString("error: unknown command; type 'help' to see available commands")
	}

	if len(args) > 1 {
		return color.YellowString("usage: help [command]")
	}

	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)

	var sb strings.Builder
	sb.WriteString(color.YellowString("Available commands:\n"))
	for _, name := range names {
		sb.WriteString(fmt.Sprintf("  %s\n", color.YellowString(commands[name])))
	}

	return strings.TrimRight(sb.String(), "\n")
}

func countDirectoryElements(root string) (int, error) {
	count := 0
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		count++
		return nil
	})
	return count, err
}

func listDirectory(root, cwd, name string) string {
	path, err := resolveSharedPath(root, cwd, name)
	if err != nil {
		return color.RedString("error: invalid path")
	}

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return color.RedString("error: not a directory")
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return color.RedString("error: could not list directory")
	}

	if len(entries) == 0 {
		return color.YellowString("(empty)")
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() {
			dirPath := filepath.Join(path, entry.Name())
			count, err := countDirectoryElements(dirPath)
			if err == nil {
				fileName = color.BlueString("%s/ (%d elements)", fileName, count)
			} else {
				fileName = color.BlueString("%s/ (?)", fileName)
			}
		} else {
			fileName = color.WhiteString("%s", fileName)
		}
		names = append(names, fileName)
	}

	total, err := countDirectoryElements(path)
	result := strings.Join(names, "    ")
	if err == nil {
		result += fmt.Sprintf("\n\n%s", color.YellowString("Total: %d elements", total))
	}

	return result
}

func readFile(root, cwd, name string, maxFileSize int64) string {
	path, err := resolveSharedPath(root, cwd, name)
	if err != nil {
		return color.RedString("error: invalid path")
	}

	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return color.RedString("error: not a regular file or file does not exist")
	}

	if info.Size() > maxFileSize {
		return color.YellowString("error: file is too large to display (maximum 64 KB)")
	}
	if info.Size() == 0{
		return color.YellowString("(empty file)")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return color.RedString("error: could not read file")
	}

	return string(data)
}

func fileInfo(root, cwd, name string) string {
	path, err := resolveSharedPath(root, cwd, name)
	if err != nil {
		return color.RedString("error: invalid path")
	}

	info, err := os.Stat(path)
	if err != nil {
		return color.RedString("error: item not found")
	}

	kind := color.WhiteString("file")
	if info.IsDir() {
		kind = color.BlueString("directory")
	}

	return fmt.Sprintf(
		"%s\n  %s %s\n  %s %s\n  %s %d bytes\n  %s %s\n  %s %s",
		color.YellowString("Item Details:"),
		color.GreenString("Name:"), info.Name(),
		color.GreenString("Type:"), kind,
		color.GreenString("Size:"), info.Size(),
		color.GreenString("Modified:"), info.ModTime().Format("2006-01-02 15:04:05"),
		color.GreenString("Permissions:"), info.Mode().Perm(),
	)
}

func treeDirectory(root, cwd, name string) string {
	path, err := resolveSharedPath(root, cwd, name)
	if err != nil {
		return color.RedString("error: invalid path")
	}

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return color.RedString("error: not a directory")
	}

	var lines []string
	if path == root {
		lines = append(lines, color.BlueString("/"))
	} else {
		lines = append(lines, color.BlueString(filepath.Base(path)+"/"))
	}

	if err := appendTree(&lines, root, path, ""); err != nil {
		return color.RedString("error: could not read directory tree")
	}

	return strings.Join(lines, "\n")
}

func appendTree(lines *[]string, root, directory, prefix string) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}

	for i, entry := range entries {
		path := filepath.Join(directory, entry.Name())
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}

		last := i == len(entries)-1
		branch := "├── "
		nextPrefix := prefix + "│   "
		if last {
			branch = "└── "
			nextPrefix = prefix + "    "
		}

		var formattedName string
		if entry.IsDir() {
			formattedName = color.BlueString("%s/", entry.Name())
		} else {
			formattedName = color.WhiteString("%s", entry.Name())
		}

		*lines = append(*lines, color.HiBlackString(prefix+branch)+formattedName)

		if entry.IsDir() {
			if err := appendTree(lines, root, path, nextPrefix); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(source, destination string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("source is not a regular file")
	}

	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, copyErr := io.Copy(dst, src)
	if copyErr != nil {
		_ = os.Remove(destination)
		return copyErr
	}
	return nil
}

func isWithinRoot(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func displayPath(cwd string) string {
	if cwd == "." || cwd == "" {
		return "Current Directory: \n" + color.BlueString("/")
	}
	return "Current Directory: \n" + color.BlueString("/"+filepath.ToSlash(cwd))
}

func cleanPath(cwd string) string {
	if cwd == "." || cwd == "" {
		return color.BlueString("/")
	}
	return color.BlueString("/" + filepath.ToSlash(cwd))
}

func resolveSharedPath(root, cwd, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty path")
	}
	if filepath.IsAbs(name) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}

	base := filepath.Clean(filepath.Join(root, cwd))
	target := filepath.Clean(filepath.Join(base, name))

	if !isWithinRoot(root, target) {
		return "", fmt.Errorf("path is outside the shared directory")
	}

	check := target
	for {
		resolved, err := filepath.EvalSymlinks(check)
		if err == nil {
			if !isWithinRoot(root, resolved) {
				return "", fmt.Errorf("symbolic link escapes shared directory")
			}
			break
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(check)
		if parent == check {
			return "", err
		}
		check = parent
	}

	return target, nil
}