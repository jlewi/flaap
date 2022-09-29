package subprocess

import (
	"bufio"
	"bytes"
	"os"
	"path"
	"sort"
	"strings"
	"testing"

	gocmd "github.com/go-cmd/cmd"
	"github.com/go-logr/zapr"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
)

func Test_RunStreaming(t *testing.T) {
	cwd, err := os.Getwd()

	if err != nil {
		t.Fatalf("Failed to get current directory; error: %v", err)
	}

	program := path.Join(cwd, "test_data", "sample.py")

	cmd := gocmd.NewCmdOptions(gocmd.Options{
		Streaming: true,
	}, "python3", program)

	log := zapr.NewLogger(zap.L())
	h := ExecHelper{Log: log}

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	statusChan := h.RunStreaming(cmd, w)

	// Wait for it to finish
	status := <-statusChan

	if status.Error != nil {
		t.Fatalf("Failed to run command; error: %v", status.Error)
	}

	expected := `stderr Line 0
stderr Line 1
stderr Line 2
stdout Line 0
stdout Line 1
stdout Line 2`
	w.Flush()
	actual := b.String()

	// Sort the actual output since there is no guarantee order is preserved when interleaving stdout and stder
	lines := strings.Split(strings.TrimSpace(actual), "\n")
	sort.Strings(lines)
	actual = strings.Join(lines, "\n")

	if d := cmp.Diff(expected, actual); d != "" {
		t.Errorf("Unexpected diff for output. Diff:\n%v", d)
	}
}
