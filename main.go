package main

import (
	"context"
	"github.com/google/go-github/v44/github" // with go modules enabled (GO111MODULE=on or outside GOPATH)
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

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

	for _, repo := range repos[:30] {
		if !repo.GetFork() {
			if repo.GetName() != user {
				log.Info().Str(repo.GetName(), repo.GetDescription()).Send()
			}
		}
	}

}
