package main

// https://medium.com/@marcus.olsson/writing-a-go-client-for-your-restful-api-c193a2f4998c

import (
	"encoding/json"
	"fmt"
	"os"

	"net/url"

	"github.com/valyala/fasthttp"
)

type config struct {
	Repositories []repository `json:"repositories"`
}

type repository struct {
	Owner  string `json:"owner"`
	Name   string `json:"name"`
	Branch string `json:"branch"`
}

type commit struct {
	Sha    string `json: "sha"`
	Author author `json: "author"`
}

type author struct {
	Login string `json: "login"`
	ID    int    `json: "id"`
}

type stats struct {
	Additions int `json: "additions"`
	Deletions int `json: "deletions"`
	Total     int `json: "total"`
}

type client struct {
	baseURL *url.URL
	http    *fasthttp.HostClient
	values  map[int]map[string]int
}

const defaultConfigFile = "config.json"
const host = "api.github.com"

func loadConfiguration(file string) (config, error) {
	var config config
	configFile, err := os.Open(file)
	if err != nil {
		return config, err
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config, nil
}

func (c *client) do(url string, v interface{}) error {
	statusCode, body, err := c.http.Get(nil, url)
	if err != nil {
		return err
	}
	if statusCode != fasthttp.StatusOK {
		return fmt.Errorf("Unexpected status code: %d. Expecting %d", statusCode, fasthttp.StatusOK)
	}

	return json.Unmarshal(body, v)
}

func (c *client) commits(repository *repository) ([]commit, error) {
	var commits []commit

	rel := &url.URL{
		Path: "repos/" + repository.Owner + "/" + repository.Name + "/commits",
	}
	u := c.baseURL.ResolveReference(rel)

	err := c.do(u.String(), &commits)
	return commits, err
}

func (c *client) stats(repository *repository, ref string) (stats, error) {
	var stats stats

	rel := &url.URL{
		Path: "repos/" + repository.Owner + "/" + repository.Name + "/commits/" + ref,
	}
	u := c.baseURL.ResolveReference(rel)

	err := c.do(u.String(), &stats)
	return stats, err
}

func (c *client) add(id int, key string, value int) {
	mm, ok := c.values[id]
	if !ok {
		mm = make(map[string]int)
		c.values[id] = mm
	}
	mm[key] += value
}

func main() {
	config, err := loadConfiguration(defaultConfigFile)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", config)

	client := &client{
		baseURL: &url.URL{
			Scheme: "https",
			Host:   host,
		},
		http: &fasthttp.HostClient{
			Addr:  host,
			IsTLS: true,
		},
	}

	commits, err2 := client.commits(&config.Repositories[0])
	if err2 != nil {
		panic(err2)
	}
	fmt.Printf("%+v\n", commits[0])
	stats, err3 := client.stats(&config.Repositories[0], commits[1].Sha)
	if err3 != nil {
		panic(err3)
	}
	fmt.Printf("%+v\n", stats)
}
