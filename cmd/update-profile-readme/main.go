package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-github/v44/github" // with go modules enabled (GO111MODULE=on or outside GOPATH)
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"os"
	"strconv"
	"strings"
	"time"
)

const MarkdownTemplate = `
<sub>**~{{COMMITS}}** commits in the last 6 months.</sub>

### Hi there 👋

⚡ Newest projects:

{{NEWEST}}
`

var Version string
var BuildTime string

func init() {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	if Version != "" {
		log.Info().Msgf("Version: %s\t", Version)
	}

	if BuildTime != "" {
		log.Info().Msgf("Build: %s\t", BuildTime)
	}

	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "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"},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	opt := &github.RepositoryListOptions{
		Sort: "created",
		Type: "public",
	}

	user := "drgomesp"
	repos, _, err := client.Repositories.List(ctx, user, opt)

	if err != nil {
		log.Fatal().Err(err).Send()
	}

	const maxRepos = 15

	newest := bytes.NewBufferString("")

	now := time.Now()
	searchResult, _, err := client.Search.Commits(
		ctx,
		fmt.Sprintf("author:drgomesp sort:date-desc committer-date:>%s",
			now.Add(time.Duration(-6)*time.Hour*24*30*6).Format("2006-01-02")),
		&github.SearchOptions{
			Sort: "author-date",
		})

	if err != nil {
		log.Fatal().Err(err).Send()
	}

	content, _, _, err := client.Repositories.GetContents(
		ctx,
		user,
		user,
		"README.md",
		nil,
	)

	_ = repos

	data, err := base64.StdEncoding.DecodeString(*content.Content)
	spew.Dump(string(data))

	for _, repo := range repos[:maxRepos] {
		if !repo.GetFork() {
			if repo.GetName() != user {
				log.Info().Str(repo.GetName(), repo.GetDescription()).Send()
				newest.WriteString(fmt.Sprintf(
					"[%s/%s](%s) %s<br/>\n",
					user,
					repo.GetName(),
					repo.GetSVNURL(),
					repo.GetDescription()),
				)
			}
		}
	}

	var tpl string
	tpl = strings.Replace(MarkdownTemplate, "{{COMMITS}}", strconv.Itoa(searchResult.GetTotal()), 1)
	tpl = strings.Replace(tpl, "{{NEWEST}}", newest.String(), 1)

	log.Info().Msg(tpl)

	// Get contents & SHA
	_, sha := getContentsAndSHA(ctx, client, "README.md", user, user)

	author := &github.CommitAuthor{Name: github.String("drgomesp"), Email: github.String("drgomesp@gmail.com")}
	_, _, err = client.Repositories.UpdateFile(ctx, user, user, "README.md", &github.RepositoryContentFileOptions{
		Message:   github.String("test"),
		SHA:       &sha,
		Content:   []byte(tpl),
		Branch:    github.String("main"),
		Author:    author,
		Committer: author,
	})

	if err != nil {
		log.Fatal().Err(err).Send()
	}
}

func getContentsAndSHA(
	ctx context.Context,
	client *github.Client,
	file string,
	owner string,
	repo string,
) (content, sha string) {
	contents, _, _, err := client.Repositories.GetContents(
		ctx,
		owner,
		repo,
		file,
		&github.RepositoryContentGetOptions{},
	)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	return *contents.Content, *contents.SHA
}
