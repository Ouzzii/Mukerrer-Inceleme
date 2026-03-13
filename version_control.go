package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/inconshreveable/go-update"
	"github.com/schollz/progressbar/v3"
)

type Release struct {
	TagName string `json:"tag_name"`
}

func doUpdate(url string) error {

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	size := resp.ContentLength

	bar := progressbar.NewOptions64(
		size,
		progressbar.OptionSetDescription("Güncelleme indiriliyor"),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionClearOnFinish(),
	)

	reader := progressbar.NewReader(resp.Body, bar)

	err = update.Apply(&reader, update.Options{})

	if err != nil {
		Closefunc(err)
		return err
	}

	fmt.Println("\nProgram güncellendi. Yeniden başlatılıyor...")
	cleanupOldExe()
	restartProgram()
	os.Exit(0)

	return nil
}

func getLatestVersion() string {
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", User, Repo))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var r Release
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(r.TagName, "v")
}

func isUploadAvailable() bool {
	latest := getLatestVersion()
	return latest != Version
}

func restartProgram() {
	exe, err := os.Executable()
	if err != nil {
		fmt.Println("Exe yolu alınamadı:", err)
		return
	}

	cmd := exec.Command(exe)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Start()

	if err != nil {
		fmt.Println("Program yeniden başlatılamadı:", err)
		return
	}

	os.Exit(0)
}

func cleanupOldExe() {
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	base := filepath.Base(exe)

	old := filepath.Join(dir, "."+base+".old")
	os.Remove(old)
}
