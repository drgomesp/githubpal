package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/thiagozs/githubpal/internal/config"
	"github.com/thiagozs/githubpal/internal/github"
	"github.com/thiagozs/githubpal/templates"
)

var (
	OrgName   string
	Repos     string
	Branch    string
	Debug     bool
	LimitList int
	FullName  string
	Name      string
	URL       string
	LoginGh   string
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the current repository with the latest changes",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		varHandlers(cmd)

		log := config.Params.GetLog().With().Str("cmd", "update").Logger()

		if config.Params.GetLogDebug() {
			log.Debug().Msgf("org: %s, repos: %s, branch: %s, debug: %t, limit: %d, full name: %s, login: %s",
				OrgName, Repos, Branch, Debug, LimitList, FullName, LoginGh)
		}

		ctx := cmd.Context()
		opts := []github.ConfigOpts{
			github.OptsToken(os.Getenv("GH_TOKEN")),
		}

		client := github.NewGithub(opts)

		if err := client.Auth(ctx); err != nil {
			log.Error().Err(err).Msg("error on auth github")
			return
		}

		owner := config.Params.GetOrgName()
		if len(config.Params.GetLoginGh()) > 0 {
			owner = config.Params.GetLoginGh()
		}

		repos, err := client.GetReposByNames(ctx, github.UPDATED, github.DESC,
			owner, config.Params.GetRepos())
		if err != nil {
			log.Error().Err(err).Msg("error on get repos")
			return
		}

		newestRepo := strings.Builder{}
		newestCommits := strings.Builder{}

		for i, repo := range repos {
			if repo.GetName() != owner {
				newestRepo.WriteString(fmt.Sprintf(
					"- **[%s/%s](%s)** %s<br/>\n",
					owner,
					repo.GetName(),
					repo.GetSVNURL(),
					repo.GetDescription()),
				)
				log.Info().Str("repo", repo.GetName()).
					Str("url", repo.GetSVNURL()).
					Str("description", repo.GetDescription()).
					Send()
			}

			if i == config.Params.GetLimitList() {
				break
			}
		}

		now := time.Now()
		weeks16 := 16 * 7 * 24 * time.Hour
		weeks16f := now.Add(-weeks16).Format("2006-01-02")
		q := fmt.Sprintf("author:%s sort:date-desc committer-date:>%s",
			config.Params.GetLoginGh(), weeks16f)

		cmts, err := client.SearchCommits(ctx, q)
		if err != nil {
			log.Error().Err(err).Msg("error on search commits")
			return
		}

		for _, cmt := range cmts {
			log.Info().Str("repo", cmt.GetRepository().GetName()).
				Str("description", cmt.GetCommit().GetMessage()).
				Str("date", cmt.GetCommit().GetCommitter().GetDate().String()).
				Send()
			newestCommits.WriteString(fmt.Sprintf(
				"- **[%s/%s](%s)** %s<br/>\n",
				owner,
				cmt.GetRepository().GetName(),
				cmt.GetRepository().GetSVNURL(),
				cmt.GetCommit().GetMessage()),
			)
		}

		var tpl string
		tpl = strings.Replace(templates.MarkdownTemplate, "{{COMMITS}}", strconv.Itoa(len(cmts)), 1)
		tpl = strings.Replace(tpl, "{{NEWEST}}", newestRepo.String(), 1)
		tpl = strings.Replace(tpl, "{{FULLNAME}}", config.Params.GetFullName(), 1)
		tpl = strings.Replace(tpl, "{{NAME}}", config.Params.GetName(), 1)
		tpl = strings.Replace(tpl, "{{URL}}", config.Params.GetURL(), -1)

		fmt.Println(tpl)

		if err := client.UpdateFileFromRepository(ctx, owner,
			owner, "main", "README.md", tpl); err != nil {
			log.Error().Err(err).Msg("error on update file")
			return
		}

	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.PersistentFlags().StringVarP(&OrgName, "org", "o", "", "organization name")
	updateCmd.PersistentFlags().StringVarP(&Repos, "repos", "r", "", "repository name")
	updateCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "debug mode")
	updateCmd.PersistentFlags().StringVarP(&Branch, "branch", "b", "", "branch name")
	updateCmd.PersistentFlags().IntVarP(&LimitList, "limit", "l", 25, "limit list")
	updateCmd.PersistentFlags().StringVarP(&FullName, "fullname", "f", "", "full name")
	updateCmd.PersistentFlags().StringVarP(&LoginGh, "login", "g", "", "login github")
	updateCmd.PersistentFlags().StringVarP(&Name, "name", "n", "", "Name")
	updateCmd.PersistentFlags().StringVarP(&URL, "url", "u", "", "URL")

}

func varHandlers(cmd *cobra.Command) error {

	opts := []config.ConfigOpts{}

	if OrgName != "" {
		opts = append(opts, config.OptsOrgName(OrgName))
	}

	if Repos != "" {
		if strings.Contains(Repos, ",") {
			opts = append(opts, config.OptsRepos(strings.Split(Repos, ",")))
		} else {
			opts = append(opts, config.OptsRepos([]string{Repos}))
		}
	}

	if Debug {
		opts = append(opts, config.OptsLogDebug(Debug))
	}

	if Branch != "" {
		opts = append(opts, config.OptsBranch(Branch))
	}

	if LimitList != 0 {
		opts = append(opts, config.OptsLimitList(LimitList))
	}

	if FullName != "" {
		opts = append(opts, config.OptsFullName(FullName))
	}

	if LoginGh != "" {
		opts = append(opts, config.OptsLoginGh(LoginGh))
	}

	if Name != "" {
		opts = append(opts, config.OptsName(Name))
	}

	if URL != "" {
		opts = append(opts, config.OptsURL(URL))
	}

	opts = append(opts, config.OptsConsoleLoggingEnabled(true))

	config.NewCfgParams(opts...)

	return nil
}
