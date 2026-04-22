package cli

import (
	"fmt"
	"strings"

	"github.com/gustavoz65/MoniMaster/internal/native"
)

type Command struct {
	Path      []string
	Args      []string
	Flags     map[string]string
	BoolFlags map[string]bool
	Raw       string
}

func Parse(input string) (Command, error) {
	tokens, err := tokenize(input)
	if err != nil {
		return Command{}, err
	}
	if len(tokens) == 0 {
		return Command{}, nil
	}
	cmd := Command{
		Flags:     map[string]string{},
		BoolFlags: map[string]bool{},
		Raw:       input,
	}

	index := 0
	for index < len(tokens) {
		token := tokens[index]
		if strings.HasPrefix(token, "--") {
			name := strings.TrimPrefix(token, "--")
			if strings.Contains(name, "=") {
				parts := strings.SplitN(name, "=", 2)
				cmd.Flags[native.NormalizeASCII(parts[0])] = parts[1]
			} else if index+1 < len(tokens) && !strings.HasPrefix(tokens[index+1], "--") {
				cmd.Flags[native.NormalizeASCII(name)] = tokens[index+1]
				index++
			} else {
				cmd.BoolFlags[native.NormalizeASCII(name)] = true
			}
		} else if len(cmd.Path) < 2 {
			cmd.Path = append(cmd.Path, native.NormalizeASCII(token))
		} else {
			cmd.Args = append(cmd.Args, token)
		}
		index++
	}
	if len(cmd.Path) == 0 && len(cmd.Args) > 0 {
		cmd.Path = append(cmd.Path, native.NormalizeASCII(cmd.Args[0]))
		cmd.Args = cmd.Args[1:]
	}
	return cmd, nil
}

func tokenize(input string) ([]string, error) {
	var (
		tokens   []string
		current  strings.Builder
		inQuotes bool
		quote    rune
	)
	for _, r := range input {
		switch {
		case (r == '"' || r == '\'') && !inQuotes:
			inQuotes = true
			quote = r
		case inQuotes && r == quote:
			inQuotes = false
		case !inQuotes && (r == ' ' || r == '\t'):
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if inQuotes {
		return nil, fmt.Errorf("aspas nao fechadas")
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens, nil
}
