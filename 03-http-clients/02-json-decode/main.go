// 02-json-decode — hit a JSON API, decode into a struct with tags.
//
// Goal: GET https://api.github.com/repos/golang/go and print
// the repo's full_name, stars, and forks count.
//
// Run:
//   go run .
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Repo models the subset of fields we care about from the GitHub API.
// Field names use `json:"..."` tags because Go's field names are
// capitalised (exported) but the JSON keys are snake_case.
type Repo struct {
	FullName        string `json:"full_name"`
	StargazersCount int    `json:"stargazers_count"`
	ForksCount      int    `json:"forks_count"`
	Description     string `json:"description,omitempty"`
}

func main() {
	const url = "https://api.github.com/repos/golang/go"

	// TODO: resp, err := http.Get(url). Handle the error. defer Body.Close().
	// TODO: guard resp.StatusCode — return an error if it isn't 200.
	// TODO: var r Repo; json.NewDecoder(resp.Body).Decode(&r). Handle error.
	// TODO: fmt.Printf the three fields.

	_ = url
	_ = http.Get
	_ = json.NewDecoder
	_ = Repo{}
	_ = fmt.Println
}
