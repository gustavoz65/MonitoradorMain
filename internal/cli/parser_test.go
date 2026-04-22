package cli

import "testing"

func TestParseCommandWithFlagsAndQuotes(t *testing.T) {
	cmd, err := Parse(`config db set --driver postgres --dsn "postgres://user:pass@localhost/db"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cmd.Path) != 2 || cmd.Path[0] != "config" || cmd.Path[1] != "db" {
		t.Fatalf("unexpected path: %#v", cmd.Path)
	}
	if len(cmd.Args) != 1 || cmd.Args[0] != "set" {
		t.Fatalf("unexpected args: %#v", cmd.Args)
	}
	if cmd.Flags["driver"] != "postgres" {
		t.Fatalf("unexpected driver flag: %q", cmd.Flags["driver"])
	}
	if cmd.Flags["dsn"] != "postgres://user:pass@localhost/db" {
		t.Fatalf("unexpected dsn flag: %q", cmd.Flags["dsn"])
	}
}
