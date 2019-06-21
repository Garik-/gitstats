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

type client struct {
	baseURL *url.URL
	http    *fasthttp.HostClient
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

func (c *client) commits(repository *repository) ([]commit, error) {
	rel := &url.URL{
		Path: "repos/" + repository.Owner + "/" + repository.Name + "/commits",
	}
	u := c.baseURL.ResolveReference(rel)

	statusCode, body, err := c.http.Get(nil, u.String())
	if err != nil {
		return nil, err
	}
	if statusCode != fasthttp.StatusOK {
		return nil, fmt.Errorf("Unexpected status code: %d. Expecting %d", statusCode, fasthttp.StatusOK)
	}

	var commits []commit
	err = json.Unmarshal(body, &commits)
	return commits, err
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
	fmt.Printf("%+v\n", commits)
}
