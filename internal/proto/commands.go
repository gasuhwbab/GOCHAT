package proto

import "strings"

type Command struct {
	Name string
	Args []string
	Raw  string
}

func IsCommand(line string) bool { return strings.HasPrefix(line, "/") }

func Parse(line string) Command {
	line = strings.TrimSpace(line)
	if line == "" {
		return Command{Name: "", Raw: line}
	}
	if !IsCommand(line) {
		return Command{Name: "", Raw: line}
	}
	fields := strings.Fields(line)
	name := strings.TrimPrefix(strings.ToLower(fields[0]), "/")
	args := []string{}
	if len(fields) > 1 {
		args = fields[1:]
	}
	return Command{Name: name, Args: args, Raw: line}
}
