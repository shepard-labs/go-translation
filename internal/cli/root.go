package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/shepard-labs/go-translation/internal/config"
	"github.com/shepard-labs/go-translation/internal/input"
	"github.com/shepard-labs/go-translation/internal/output"
	"github.com/shepard-labs/go-translation/internal/providers"
	"github.com/shepard-labs/go-translation/internal/translate"
	"github.com/shepard-labs/go-translation/internal/tui"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	provider        string
	model           string
	from            string
	to              string
	configPath      string
	timeout         time.Duration
	maxOutputTokens int
	temperature     float64
	temperatureSet  bool
	text            string
	file            string
	out             string
	force           bool
	verbose         bool
}

func Execute() error {
	return NewRootCommand().Execute()
}

func NewRootCommand() *cobra.Command {
	opts := &rootOptions{}
	cmd := &cobra.Command{
		Use:   "gotranslate [text]",
		Short: "Translate text and files with local CLI access to AI providers",
		Long:  "gotranslate is a local-only CLI for AI-assisted translation using Anthropic, Gemini, OpenRouter, or Google Cloud.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTranslate(cmd.Context(), cmd, opts, args)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.provider, "provider", "", "Provider to use: anthropic, gemini, openrouter, or googlecloud")
	flags.StringVar(&opts.model, "model", "", "Model to use")
	flags.StringVar(&opts.from, "from", "", "Source language, or auto")
	flags.StringVar(&opts.to, "to", "", "Target language")
	flags.StringVar(&opts.configPath, "config", "", "Config file path")
	flags.DurationVar(&opts.timeout, "timeout", 0, "Request timeout")
	flags.IntVar(&opts.maxOutputTokens, "max-output-tokens", 0, "Maximum output tokens")
	flags.Float64Var(&opts.temperature, "temperature", 0, "Sampling temperature")
	flags.StringVar(&opts.text, "text", "", "Text to translate")
	flags.StringVar(&opts.file, "file", "", "File to translate")
	flags.StringVar(&opts.out, "out", "", "Output file path")
	flags.BoolVar(&opts.force, "force", false, "Overwrite existing output/config files")
	flags.BoolVar(&opts.verbose, "verbose", false, "Print extra status information")
	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		opts.temperatureSet = cmd.Flags().Changed("temperature")
	}

	cmd.AddCommand(newConfigCommand(opts, cmd))
	cmd.AddCommand(newProvidersCommand())
	cmd.AddCommand(newModelsCommand())
	cmd.AddCommand(newTUICommand(opts, cmd))
	return cmd
}

func runTranslate(ctx context.Context, cmd *cobra.Command, opts *rootOptions, args []string) error {
	cfg, _, err := loadResolvedConfig(cmd, opts)
	if err != nil {
		return withCode(ExitUsage, err)
	}
	if err := config.Validate(cfg); err != nil {
		var missing config.MissingCredentialsError
		if errors.As(err, &missing) {
			return withCode(ExitMissingCredentials, err)
		}
		return withCode(ExitUsage, err)
	}
	text, err := input.Read(input.Options{Args: args, Text: opts.text, File: opts.file, Stdin: os.Stdin})
	if err != nil {
		return withCode(ExitIO, err)
	}
	translator, err := providers.New(cfg)
	if err != nil {
		return withCode(ExitUsage, err)
	}
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	service := translate.NewService(translator, cfg.Chunking.Enabled, cfg.Chunking.MaxChars)
	result, err := service.Translate(ctx, translate.TranslationRequest{
		Provider:        cfg.DefaultProvider,
		Model:           config.ModelFor(cfg),
		SourceLanguage:  cfg.DefaultSourceLanguage,
		TargetLanguage:  cfg.DefaultTargetLanguage,
		Text:            text,
		MaxOutputTokens: cfg.MaxOutputTokens,
		Temperature:     cfg.Temperature,
	})
	if err != nil {
		return withCode(ExitRuntime, err)
	}
	if err := output.Write(result.Text, opts.out, opts.force, os.Stdout); err != nil {
		return withCode(ExitIO, err)
	}
	return nil
}

func loadResolvedConfig(cmd *cobra.Command, opts *rootOptions) (config.Config, string, error) {
	var temp *float64
	if opts.temperatureSet {
		temp = &opts.temperature
	}
	cfg, path, err := config.LoadWithLoader(opts.configPath, config.Overrides{
		Provider:        opts.provider,
		Model:           opts.model,
		SourceLanguage:  opts.from,
		TargetLanguage:  opts.to,
		Timeout:         opts.timeout,
		MaxOutputTokens: opts.maxOutputTokens,
		Temperature:     temp,
	})
	if err != nil {
		return cfg, path, err
	}
	return cfg, path, nil
}

func newTUICommand(opts *rootOptions, root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Open the interactive translation TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadResolvedConfig(cmd, opts)
			if err != nil {
				return withCode(ExitUsage, err)
			}
			if err := tui.Run(cmd.Context(), cfg); err != nil {
				return withCode(ExitRuntime, err)
			}
			return nil
		},
	}
	cmd.Flags().AddFlagSet(root.Flags())
	return cmd
}

func printProviderStatus(cfg config.Config) string {
	provider := cfg.Providers[cfg.DefaultProvider]
	set := "no"
	if os.Getenv(provider.APIKeyEnv) != "" {
		set = "yes"
	}
	return fmt.Sprintf("Provider: %s\nModel: %s\nAPI key env: %s\nAPI key set: %s\n", cfg.DefaultProvider, config.ModelFor(cfg), provider.APIKeyEnv, set)
}
