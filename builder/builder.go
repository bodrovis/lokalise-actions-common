package builder

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Options struct {
	SourceRoot string   // e.g. "src"
	OutputDir  string   // e.g. "bin"
	Binaries   []string // list of subfolder names
	Compress   bool     // use UPX
	Build      bool     // build binaries
	Lint       bool     // run gofmt + gofumpt
}

func Run(opts Options) error {
	root := getProjectRoot()
	binDir := filepath.Join(root, opts.OutputDir)

	if err := os.MkdirAll(binDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create bin dir: %w", err)
	}

	// Ensure gofumpt is installed
	if !checkCommand("gofumpt") {
		if err := runCommand("go", []string{"install", "mvdan.cc/gofumpt@latest"}); err != nil {
			fmt.Println("warning: could not install gofumpt:", err)
		}
	}

	for _, name := range opts.Binaries {
		pkgPath := filepath.Join(root, opts.SourceRoot, name)

		if opts.Lint {
			if err := lint(pkgPath); err != nil {
				fmt.Println("lint failed:", err)
			}
		}

		if opts.Build {
			built, err := buildBinary(pkgPath, binDir, name)
			if err != nil {
				log.Printf("Build failed for %s: %v", name, err)
				continue
			}

			if opts.Compress && checkCommand("upx") {
				for _, bin := range built {
					if isLinuxBinary(bin) {
						_ = compressWithUPX(bin) // ignore error, log inside
					}
				}
			}
		}
	}

	return nil
}

// -------- helpers --------

func lint(dir string) error {
	fmt.Printf("Linting %s...\n", dir)

	cmd := exec.Command("go", "fmt", "./...")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go fmt error: %w", err)
	}

	if err := runCommand("gofumpt", []string{"-l", "-w", dir}); err != nil {
		return fmt.Errorf("gofumpt error: %w", err)
	}
	return nil
}

func buildBinary(srcDir, outputDir, binaryName string) ([]string, error) {
	targets := []struct {
		goos, goarch, suffix string
	}{
		{"linux", "amd64", "_linux_amd64"},
		{"linux", "arm64", "_linux_arm64"},
		{"darwin", "amd64", "_mac_amd64"},
		{"darwin", "arm64", "_mac_arm64"},
	}

	var paths []string
	for _, t := range targets {
		out := filepath.Join(outputDir, binaryName+t.suffix)
		fmt.Printf("Building %s for %s/%s...\n", binaryName, t.goos, t.goarch)

		ldflags := []string{
			"-s", "-w",
			"-buildid=",
			"-extldflags=-static",
		}
		cmd := exec.Command("go", "build",
			"-tags=netgo,osusergo",
			"-trimpath",
			"-ldflags", strings.Join(ldflags, " "),
			"-o", out,
		)
		cmd.Dir = srcDir
		cmd.Env = append(os.Environ(),
			"GOOS="+t.goos,
			"GOARCH="+t.goarch,
			"CGO_ENABLED=0",
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to build for %s/%s: %w", t.goos, t.goarch, err)
		}
		if checkCommand("strip") {
			fmt.Println("Stripping binary:", out)
			if err := exec.Command("strip", out).Run(); err != nil {
				log.Printf("strip failed for %s: %v", out, err)
			}
		}
		paths = append(paths, out)
	}
	return paths, nil
}

func compressWithUPX(path string) error {
	fmt.Println("Compressing with UPX:", path)
	err := runCommand("upx", []string{"--best", "--lzma", path})
	if err != nil {
		log.Printf("UPX compression failed for %s: %v", path, err)
	}
	return err
}

func isLinuxBinary(path string) bool {
	return filepath.Ext(path) == "" &&
		(strings.HasSuffix(path, "_linux_amd64") || strings.HasSuffix(path, "_linux_arm64"))
}

func runCommand(cmd string, args []string) error {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func checkCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func getProjectRoot() string {
	root, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get project root: %v", err)
	}
	return root
}
