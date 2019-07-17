package main

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"testing"
)

var mainTests = []struct {
	args   []string
	stdin  string
	stdout string
	stderr string
}{
	{[]string{"true"}, "", "", ""},
	{[]string{"false"}, "", "", "exit status 1\n"},
	{[]string{"sh", "-c", "echo -n Hello, world!"}, "", "", ""},
	{[]string{"sh", "-c", "sleep 2 && echo -n Hello, world!"}, "", "Hello, world!", ""},
}

func TestTrue(t *testing.T) {

	for _, tt := range mainTests {
		t.Run(fmt.Sprintf("command %s", tt.args), func(t *testing.T) {
			//			t.Parallel()

			args := []string{"go", "run", "main.go", "1"}
			args = append(args, tt.args...)

			cmd := exec.Command(args[0], args[1:]...)
			stdin, err := cmd.StdinPipe()
			if err != nil {
				t.Errorf("failed to prepare command input: %s", err)
			}
			io.WriteString(stdin, tt.stdin)
			stdin.Close()

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			cmd.Run()

			output := stdout.String()
			if strings.Compare(output, tt.stdout) != 0 {
				t.Errorf("execution stdout (%s) => %q, want %q", tt.args, output, tt.stdout)
			}
			errors := stderr.String()
			if strings.Compare(errors, tt.stderr) != 0 {
				t.Errorf("execution stderr (%s) => %q, want %q", tt.args, errors, tt.stderr)
			}
		})
	}
}
