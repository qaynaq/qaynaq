package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"

	_ "github.com/qaynaq/qaynaq/internal/components/all"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	_ "github.com/warpstreamlabs/bento/public/components/all"
)

var (
	Version   = "dev"
	DateBuilt = "unknown"
)

const (
	RoleCoordinator = "coordinator"
	RoleWorker      = "worker"
)

func main() {
	app := &cli.App{
		Name:    "run",
		Usage:   "Run the qaynaq server",
		Suggest: true,
		Version: Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
			},
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "role",
				Aliases: []string{"r"},
				Usage:   "role of the node (coordinator or worker)",
				EnvVars: []string{"ROLE"},
				Value:   RoleCoordinator,
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "discovery-uri",
				Aliases: []string{"du"},
				Usage:   "node discovery URI",
				EnvVars: []string{"DISCOVERY_URI"},
				Value:   "localhost:50000",
			}),
			altsrc.NewUintFlag(&cli.UintFlag{
				Name:    "http-port",
				Aliases: []string{"hp"},
				Usage:   "http port of the node",
				EnvVars: []string{"HTTP_PORT"},
				Value:   8080,
			}),
			altsrc.NewUintFlag(&cli.UintFlag{
				Name:    "grpc-port",
				Aliases: []string{"gp"},
				Usage:   "grpc port of the node",
				EnvVars: []string{"GRPC_PORT"},
				Value:   50000,
			}),
			altsrc.NewBoolFlag(&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "debug mode",
				EnvVars: []string{"DEBUG_MODE"},
				Value:   false,
			}),
			// Database
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "database.driver",
				Usage:   "database driver (sqlite or postgres)",
				EnvVars: []string{"DATABASE_DRIVER"},
				Value:   "sqlite",
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "database.uri",
				Usage:   "database connection URI",
				EnvVars: []string{"DATABASE_URI"},
			}),
			// Secret
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "secret.key",
				Usage:   "encryption key (must be exactly 32 bytes)",
				EnvVars: []string{"SECRET_KEY"},
			}),
			// Auth
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "auth.type",
				Usage:   "authentication type (none, basic, or oauth2)",
				EnvVars: []string{"AUTH_TYPE"},
				Value:   "none",
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "auth.basic-username",
				Usage:   "basic auth username",
				EnvVars: []string{"AUTH_BASIC_USERNAME"},
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "auth.basic-password",
				Usage:   "basic auth password",
				EnvVars: []string{"AUTH_BASIC_PASSWORD"},
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "auth.oauth2-client-id",
				Usage:   "OAuth2 client ID",
				EnvVars: []string{"AUTH_OAUTH2_CLIENT_ID"},
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "auth.oauth2-client-secret",
				Usage:   "OAuth2 client secret",
				EnvVars: []string{"AUTH_OAUTH2_CLIENT_SECRET"},
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "auth.oauth2-authorization-url",
				Usage:   "OAuth2 authorization endpoint",
				EnvVars: []string{"AUTH_OAUTH2_AUTHORIZATION_URL"},
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "auth.oauth2-token-url",
				Usage:   "OAuth2 token endpoint",
				EnvVars: []string{"AUTH_OAUTH2_TOKEN_URL"},
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "auth.oauth2-redirect-url",
				Usage:   "OAuth2 redirect/callback URL",
				EnvVars: []string{"AUTH_OAUTH2_REDIRECT_URL"},
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "auth.oauth2-scopes",
				Usage:   "OAuth2 scopes (comma-separated)",
				EnvVars: []string{"AUTH_OAUTH2_SCOPES"},
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "auth.oauth2-user-info-url",
				Usage:   "OAuth2 user info endpoint",
				EnvVars: []string{"AUTH_OAUTH2_USER_INFO_URL"},
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "auth.oauth2-allowed-users",
				Usage:   "OAuth2 allowed users (comma-separated)",
				EnvVars: []string{"AUTH_OAUTH2_ALLOWED_USERS"},
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "auth.oauth2-allowed-domains",
				Usage:   "OAuth2 allowed domains (comma-separated)",
				EnvVars: []string{"AUTH_OAUTH2_ALLOWED_DOMAINS"},
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "auth.oauth2-session-cookie-name",
				Usage:   "OAuth2 session cookie name",
				EnvVars: []string{"AUTH_OAUTH2_SESSION_COOKIE_NAME"},
				Value:   "qaynaq_session",
			}),
		},
		Before: func(ctx *cli.Context) error {
			configFile := ctx.String("config")
			if configFile != "" {
				log.Info().Str("config-file", configFile).Msg("Attempting to load configuration from file")
				yamlSource, err := altsrc.NewYamlSourceFromFile(configFile)
				if err != nil {
					return fmt.Errorf("failed to load config file %s: %w", configFile, err)
				}
				err = altsrc.ApplyInputSourceValues(ctx, yamlSource, ctx.App.Flags)
				if err != nil {
					return fmt.Errorf("error applying config source values: %w", err)
				}
				log.Info().Msg("Successfully loaded configuration from file")
			}

			role := ctx.String("role")
			if role != RoleCoordinator && role != RoleWorker {
				return fmt.Errorf("invalid role: %s. Must be '%s' or '%s'", role, RoleCoordinator, RoleWorker)
			}

			if ctx.String("secret.key") == "" {
				return fmt.Errorf("secret.key is required (set via YAML, SECRET_KEY env, or --secret.key flag)")
			}

			if role == RoleCoordinator && ctx.String("database.uri") == "" {
				return fmt.Errorf("database.uri is required for coordinator (set via YAML, DATABASE_URI env, or --database.uri flag)")
			}

			return nil
		},
		Action: func(ctx *cli.Context) error {
			cCtx, stop := signal.NotifyContext(ctx.Context, os.Interrupt, os.Kill)
			defer stop()
			setLogLevel(ctx.Bool("debug"))
			if ctx.String("role") == RoleCoordinator {
				log.Info().Msg("starting coordinator")
				coordinatorCLI := InitializeCoordinatorCommand(ctx)
				coordinatorCLI.Run(cCtx)
			} else {
				log.Info().Msg("starting worker")
				workerCLI := InitializeWorkerCommand(cCtx, ctx)
				workerCLI.Run(cCtx)
			}
			log.Info().Msg("shutting node down")
			return nil
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:  "hello",
			Usage: "Open the setup wizard in your browser",
			Flags: []cli.Flag{
				&cli.UintFlag{
					Name:  "port",
					Usage: "HTTP port of the coordinator",
					Value: 8080,
				},
			},
			Action: func(ctx *cli.Context) error {
				url := fmt.Sprintf("http://localhost:%d", ctx.Uint("port"))
				fmt.Printf("Opening %s ...\n", url)
				return openBrowser(url)
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("unable to start qaynaq")
	}
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("cmd", "/c", "start", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}

func setLogLevel(debug bool) {
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
