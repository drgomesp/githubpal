package config

import (
	"io"
	"os"
	"path"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Params *CfgParams

type ConfigOpts func(c *CfgParams) error

type CfgParams struct {
	Version               string
	Debug                 bool
	DisableLog            bool
	ConsoleLoggingEnabled bool
	FileLoggingEnabled    bool
	FileLoggingName       string
	OrgName               string
	Repos                 []string
	Branch                string
	LimitList             int
	FullName              string
	Name                  string
	LoginGh               string
	URL                   string
	Log                   zerolog.Logger
}

type ConfigLogger struct {
	Directory  string
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

func NewCfgParams(opts ...ConfigOpts) *CfgParams {
	params := &CfgParams{}
	for _, opt := range opts {
		if err := opt(params); err != nil {
			panic(err)
		}
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	var writers []io.Writer

	if params.GetConsoleLoggingEnabled() {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if params.GetFileLoggingEnabled() {
		writers = append(writers, newRollingFile(ConfigLogger{
			Directory:  "logs",
			Filename:   params.GetFileLoggingName(), // file.log
			MaxSize:    10,                          // megabytes
			MaxBackups: 3,                           // Total files
			MaxAge:     28,                          // days
		}))
	}

	mw := io.MultiWriter(writers...)

	log := zerolog.New(mw).With().Timestamp().Logger()

	if params.GetLogDebug() {
		log.Info().Bool("debug", true).Msg("log debug")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if params.GetLogDisable() {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}

	params.Log = log
	Params = params

	return params
}

func newRollingFile(config ConfigLogger) io.Writer {
	if err := os.MkdirAll(config.Directory, 0744); err != nil {
		log.Error().Err(err).Str("path", config.Directory).Msg("can't create log directory")
		return nil
	}

	return &lumberjack.Logger{
		Filename:   path.Join(config.Directory, config.Filename),
		MaxBackups: config.MaxBackups, // files
		MaxSize:    config.MaxSize,    // megabytes
		MaxAge:     config.MaxAge,     // days
	}
}

func OptsVersion(version string) ConfigOpts {
	return func(c *CfgParams) error {
		c.Version = version
		return nil
	}
}

func OptsLogDebug(debug bool) ConfigOpts {
	return func(c *CfgParams) error {
		c.Debug = debug
		return nil
	}
}

func OptsLogDisable(disable bool) ConfigOpts {
	return func(c *CfgParams) error {
		c.DisableLog = disable
		return nil
	}
}

func OptsConsoleLoggingEnabled(enabled bool) ConfigOpts {
	return func(c *CfgParams) error {
		c.ConsoleLoggingEnabled = enabled
		return nil
	}
}

func OptsFileLoggingEnabled(enabled bool) ConfigOpts {
	return func(c *CfgParams) error {
		c.FileLoggingEnabled = enabled
		return nil
	}
}

func OptsFileLoggingName(name string) ConfigOpts {
	return func(c *CfgParams) error {
		c.FileLoggingName = name
		return nil
	}
}

func OptsOrgName(name string) ConfigOpts {
	return func(c *CfgParams) error {
		c.OrgName = name
		return nil
	}
}

func OptsRepos(repos []string) ConfigOpts {
	return func(c *CfgParams) error {
		c.Repos = repos
		return nil
	}
}

func OptsBranch(branch string) ConfigOpts {
	return func(c *CfgParams) error {
		c.Branch = branch
		return nil
	}
}

func OptsLimitList(limit int) ConfigOpts {
	return func(c *CfgParams) error {
		c.LimitList = limit
		return nil
	}
}

func OptsFullName(name string) ConfigOpts {
	return func(c *CfgParams) error {
		c.FullName = name
		return nil
	}
}

func OptsLoginGh(login string) ConfigOpts {
	return func(c *CfgParams) error {
		c.LoginGh = login
		return nil
	}
}

func OptsName(name string) ConfigOpts {
	return func(c *CfgParams) error {
		c.Name = name
		return nil
	}
}

func OptsURL(url string) ConfigOpts {
	return func(c *CfgParams) error {
		c.URL = url
		return nil
	}
}

// ------------- getters

func (c *CfgParams) GetVersion() string {
	return c.Version
}

func (c *CfgParams) GetLogDebug() bool {
	return c.Debug
}

func (c *CfgParams) GetLogDisable() bool {
	return c.DisableLog
}

func (c *CfgParams) GetLog() zerolog.Logger {
	return c.Log
}

func (c *CfgParams) GetConsoleLoggingEnabled() bool {
	return c.ConsoleLoggingEnabled
}

func (c *CfgParams) GetFileLoggingEnabled() bool {
	return c.FileLoggingEnabled
}

func (c *CfgParams) GetFileLoggingName() string {
	return c.FileLoggingName
}

func (c *CfgParams) GetOrgName() string {
	return c.OrgName
}

func (c *CfgParams) GetRepos() []string {
	return c.Repos
}

func (c *CfgParams) GetBranch() string {
	return c.Branch
}

func (c *CfgParams) GetLimitList() int {
	return c.LimitList
}

func (c *CfgParams) GetFullName() string {
	return c.FullName
}

func (c *CfgParams) GetLoginGh() string {
	return c.LoginGh
}

func (c *CfgParams) GetName() string {
	return c.Name
}

func (c *CfgParams) GetURL() string {
	return c.URL
}
