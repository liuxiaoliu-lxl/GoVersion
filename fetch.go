package main

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"strings"
)

func fetchGoVersions() ([]GoVersion, error) {
	resp, err := http.Get("https://go.dev/dl/?mode=json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var versions []GoVersion
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, err
	}

	return versions, nil
}

func fetchGoVersionsFromWeb() ([]GoVersion, error) {
	resp, err := http.Get("https://go.dev/dl/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析 HTML 页面
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var versions []GoVersion
	// Find all versions in the page
	doc.Find("div.toggleVisible").Each(func(i int, s *goquery.Selection) {
		version := s.AttrOr("id", "")
		if version == "" {
			return
		}

		var files []GoFile

		// Parse the download table for this version
		s.Find(".downloadtable tr").Each(func(i int, tr *goquery.Selection) {
			// Skip rows that don't contain files
			if tr.HasClass("first") || tr.HasClass("js-togglePorts") {
				return
			}

			fileName := tr.Find(".filename a").Text()
			if fileName == "" {
				return
			}

			kind := tr.Find("td").Eq(1).Text()
			os := tr.Find("td").Eq(2).Text()
			arch := tr.Find("td").Eq(3).Text()

			// Only add files that are archives
			if os != "" && kind == "Archive" {
				files = append(files, GoFile{
					Filename: fileName,
					OS:       os,
					Arch:     arch,
					Version:  version,
					Kind:     kind,
				})
			}
		})

		// Append the version with its files to the versions list
		versions = append(versions, GoVersion{
			Version: version,
			Stable:  strings.Contains(version, "go1"),
			Files:   files,
		})
	})

	// 提取所有版本的信息
	doc.Find(".toggle").Each(func(i int, s *goquery.Selection) {
		version := s.AttrOr("id", "")
		if version == "" {
			return
		}

		var files []GoFile
		s.Find(".downloadtable tr").Each(func(i int, tr *goquery.Selection) {
			fileName := tr.Find(".filename a").Text()
			if fileName == "" {
				return
			}

			kind := tr.Find("td").Eq(1).Text()
			os := tr.Find("td").Eq(2).Text()
			arch := tr.Find("td").Eq(3).Text()

			if os != "" {
				files = append(files, GoFile{
					Filename: fileName,
					OS:       os,
					Arch:     arch,
					Version:  version,
					Kind:     kind,
				})
			}
		})

		versions = append(versions, GoVersion{
			Version: version,
			Stable:  strings.Contains(version, "go1"),
			Files:   files,
		})
	})

	return versions, nil
}
