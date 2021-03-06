// pkgcritic is a tool combining search results from GoDoc API and stats info from GitHub API.
//
// pkgcritic also uses indentation to show the fork relationship between packages.
// The GoDoc search result is re-ordered by GitHub Stars.
//
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	debugpkg "runtime/debug"
	"sort"
	"strings"
	"sync"

	"github.com/google/go-github/github"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

var debug bool
var token string

func init() {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)
}

func main() {
	flag.BoolVar(&debug, "debug", false, "print debug info")
	web := flag.Bool("web", false, "show result in browser and start a web server")
	open := flag.Bool("open", false, "open browser")
	query := flag.String("q", "", "godoc query keyword")
	port := flag.String("port", "9090", "web server port number")
	flag.StringVar(&token, "github-token", "", "github token for more github requests")
	flag.Parse()

	if token == "" {
		fmt.Println("specify -github-token can increase github api rate limits: https://github.com/settings/tokens")
	}

	if *web {
		type data struct {
			Query               string
			GitHubs, NonGitHubs []*Critique
		}
		caches := map[string]data{}
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			query := strings.TrimSpace(r.URL.Query().Get("query"))
			if query != "" {
				log.Println("Query:", query)
			}
			d, ok := caches[query]
			if ok {
				goto render
			} else if query != "" {
				githubs, nonGithubs, err := report(query)
				if err != nil {
					log.Println(err)
					if _, err := w.Write([]byte(err.Error() + "\n" + string(debugpkg.Stack()))); err != nil {
						log.Println(err)
					}
					debugpkg.PrintStack()
					return
				}
				d = data{
					Query:      query,
					GitHubs:    githubs,
					NonGitHubs: nonGithubs,
				}
				caches[query] = d
			}

		render:
			if err := tmpl.ExecuteTemplate(w, "main", d); err != nil {
				log.Println(err)
			}
		})

		log.Println("Listening on :" + *port)
		if *open {
			if err := browser.OpenURL("http://localhost:" + *port); err != nil {
				log.Println(err)
			}
		}
		log.Fatal(http.ListenAndServe(":"+*port, nil))
	} else {
		if *query == "" {
			log.Println("Please specify query keyword by -q")
			os.Exit(1)
		}
		githubs, nonGithubs, err := report(*query)
		if err != nil {
			exit(err)
		}

		for i, pkgs := range [][]*Critique{githubs, nonGithubs} {
			fmt.Println("===========================")
			if i == 0 {
				fmt.Println("GitHub Packages")
			} else {
				fmt.Println("Non-GitHub Packages")
			}
			printPkgs(pkgs, "")
		}
	}
}

func report(query string) (githubs, nonGithubs []*Critique, err error) {
	resp, err := http.Get("http://api.godoc.org/search?q=" + url.QueryEscape(query))
	if err != nil {
		return
	}

	var results []*Critique
	if err = json.NewDecoder(resp.Body).Decode(&struct{ Results *[]*Critique }{&results}); err != nil {
		return
	}

	var tc *http.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc = oauth2.NewClient(oauth2.NoContext, ts)
	}

	client := github.NewClient(tc)

	var wg sync.WaitGroup
	donechan := make(chan *Critique)
	jobchan := make(chan *Critique)
	rlchan := make(chan *Critique, 10)

	go func() {
		for _, r := range results {
			if !strings.HasPrefix(r.Path, "github.com") {
				nonGithubs = append(nonGithubs, r)
				continue
			}

			wg.Add(1)
			rlchan <- r
			jobchan <- r
		}
		close(jobchan)
	}()

	go func() {
		for c := range donechan {
			githubs = append(githubs, c)
		}
		if debug {
			log.Println("received finished")
		}
	}()

	for {
		job, ok := <-jobchan
		if !ok {
			// close(donechan)
			break
		}
		go func(r *Critique) {
			defer wg.Done()
			defer func() { <-rlchan }()

			if debug {
				log.Println("handling:", r.Path)
			}
			parts := strings.Split(r.Path, "/")
			owner, repo := parts[1], parts[2]
			var c Critique
			c.GitHubFullName = strings.Join([]string{owner, repo}, "/")
			c.Path = r.Path
			c.Synopsis = r.Synopsis
			c.Score = r.Score

			if c.Repository, _, err = client.Repositories.Get(owner, repo); err != nil {
				log.Println("error", c.Path, err)
				return
			}
			forks, _, err := client.Repositories.ListForks(owner, repo, nil)
			if err != nil {
				log.Println("error", c.Path, err)
				return
			}
			c.forks = forks
			donechan <- &c
			if debug {
				log.Println("completed:", r.Path)
			}
		}(job)
	}

	wg.Wait()
	close(donechan)
	close(rlchan)

	if debug {
		log.Println("hierarchise")
	}
	hierarchise(githubs)

	var topLevels []*Critique
	for _, c := range githubs {
		if c.forker {
			continue
		}
		topLevels = append(topLevels, c)
	}

	githubs = topLevels
	sort.Sort(ByStar(githubs))

	return
}

type Critique struct {
	Path     string
	Synopsis string
	Score    float64

	GitHubFullName string
	Forks          []*Critique

	*github.Repository
	// Stargazers     int

	forks  []github.Repository
	forker bool
}

func exit(err error) {
	log.Println(err)
	debugpkg.PrintStack()
	os.Exit(1)
}

func hierarchise(cs []*Critique) {
	for _, c1 := range cs {
		for _, repo := range c1.forks {
			for _, c2 := range cs {
				if *repo.FullName != c2.GitHubFullName {
					continue
				}
				c2.forker = true
				c1.Forks = append(c1.Forks, c2)
			}
		}
	}
}

type ByStar []*Critique

func (cs ByStar) Len() int {
	return len(cs)
}

func (cs ByStar) Less(i int, j int) bool {
	return *cs[i].StargazersCount > *cs[j].StargazersCount
}

func (cs ByStar) Swap(i int, j int) {
	tmp := cs[i]
	cs[i] = cs[j]
	cs[j] = tmp
}

func printPkgs(pkgs []*Critique, prefix string) {
	for _, pkg := range pkgs {
		fmt.Println(prefix + "")
		fmt.Println(prefix + pkg.Path)
		if pkg.Repository != nil {
			fmt.Printf(prefix+"Stars: %d Forks: %d UpdatedAt: %s CreatedAt: %s\n", *pkg.StargazersCount, *pkg.ForksCount, pkg.UpdatedAt.Format("2006-01-02"), pkg.CreatedAt.Format("2006-01-02"))
		}
		if pkg.Synopsis != "" {
			fmt.Println(prefix + pkg.Synopsis)
		}
		if len(pkg.Forks) > 0 {
			printPkgs(pkg.Forks, prefix+"    ")
		}
	}
}
