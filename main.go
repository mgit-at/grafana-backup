// Copyright (c) 2015 by Christoph Hack <chack@mgit.at>

// grafana-backup stores all Grafana dashboards locally in JSON files, useful
// for submitting into a git repository.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

func get(apiRoot, apiPath, apiKey string, result interface{}) error {
	u, err := url.ParseRequestURI(apiRoot)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, "api", apiPath)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(result)
}

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s <rootURL> <apiKey>\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	apiRoot := flag.Arg(0)
	apiKey := flag.Arg(1)

	var list []struct {
		URI string `json:"uri"`
	}
	if err := get(apiRoot, "/search/", apiKey, &list); err != nil {
		log.Fatalln("failed to search for dashboards:", err)
	}
	for i := range list {
		name := list[i].URI
		fmt.Println("Exporting", path.Base(name))
		var data json.RawMessage
		if err := get(apiRoot, path.Join("/dashboards/", name), apiKey, &data); err != nil {
			log.Fatalln("failed to get dashboard:", err)
		}
		buf := &bytes.Buffer{}
		json.Indent(buf, data, "", "  ")
		if err := ioutil.WriteFile(fmt.Sprintf("%s.json", path.Base(name)), buf.Bytes(), 0644); err != nil {
			log.Fatalln("failed to write file:", err)
		}
	}
}
