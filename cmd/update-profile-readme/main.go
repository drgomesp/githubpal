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
	"os"
	"strings"
)

const MarkdownTemplate = `
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
	ctx := context.Background()

	if Version != "" {
		log.Info().Msgf("Version: %s\t", Version)
	}

	if BuildTime != "" {
		log.Info().Msgf("Build: %s\t", BuildTime)
	}

	client := github.NewClient(nil)

	opt := &github.RepositoryListOptions{
		Sort: "created",
		Type: "public",
	}

	user := "drgomesp"
	repos, _, err := client.Repositories.List(context.Background(), user, opt)

	if err != nil {
		log.Fatal().Err(err).Send()
	}

	const maxRepos = 20

	newest := bytes.NewBufferString("")
	for _, repo := range repos[:maxRepos] {
		if !repo.GetFork() {
			if repo.GetName() != user {
				log.Debug().Str(repo.GetName(), repo.GetDescription()).Send()
				newest.WriteString(fmt.Sprintf("[%s/%s](%s) %s<br/>\n", user, repo.GetName(), repo.GetURL(), repo.GetDescription()))
			}
		}
	}

	content, _, _, err := client.Repositories.GetContents(
		ctx,
		user,
		"drgomesp",
		"README.md",
		nil,
	)

	_ = repos

	data, err := base64.StdEncoding.DecodeString(*content.Content)
	spew.Dump(string(data))

	tpl := strings.Replace(MarkdownTemplate, "{{NEWEST}}", newest.String(), 1)

	log.Info().Msg(tpl)
}
