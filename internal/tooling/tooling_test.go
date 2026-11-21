package tooling

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestStaticLinuxCrossCompilation(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	for _, arch := range []string{"arm64", "amd64"} {
		t.Run(arch, func(t *testing.T) {
			cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", filepath.Join(t.TempDir(), "buildprobe"), "./testdata/tooling/buildprobe")
			cmd.Dir = root
			cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=linux", "GOARCH="+arch)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("cross-compile %s: %v\n%s", arch, err, output)
			}
		})
	}
}
