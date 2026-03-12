package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/inconshreveable/go-update"
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
	err = update.Apply(resp.Body, update.Options{})
	if err != nil {
		// error handling
	}
	return err
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
