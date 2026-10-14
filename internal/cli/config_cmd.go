package cli

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/shepard-labs/go-translation/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCommand(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "config", Short: "Manage local configuration"}
	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Create a local config file with guided prompts",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Default()
			path := opts.configPath
			if path == "" {
				var err error
				path, err = config.DefaultPath()
				if err != nil {
					return withCode(ExitIO, err)
				}
			}
			if err := runConfigForm(&cfg); err != nil {
				return withCode(ExitUsage, err)
			}
			if err := config.Write(path, cfg, opts.force); err != nil {
				return withCode(ExitIO, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Wrote config to %s\n", path)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show resolved local configuration without secrets",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := loadResolvedConfig(opts)
			if err != nil {
				return withCode(ExitUsage, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Config: %s\n%s", path, printProviderStatus(cfg))
			fmt.Fprintf(cmd.OutOrStdout(), "From: %s\nTo: %s\nTimeout: %s\nMax output tokens: %d\nChunking: %t (%d chars)\n", cfg.DefaultSourceLanguage, cfg.DefaultTargetLanguage, cfg.Timeout, cfg.MaxOutputTokens, cfg.Chunking.Enabled, cfg.Chunking.MaxChars)
			return nil
		},
	})
	return cmd
}

func runConfigForm(cfg *config.Config) error {
	anthropic := cfg.Providers[config.ProviderAnthropic]
	gemini := cfg.Providers[config.ProviderGemini]
	openrouter := cfg.Providers[config.ProviderOpenRouter]
	googlecloud := cfg.Providers[config.ProviderGoogleCloud]
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().Title("Default provider").Options(
			huh.NewOption("Anthropic", config.ProviderAnthropic),
			huh.NewOption("Gemini", config.ProviderGemini),
			huh.NewOption("OpenRouter", config.ProviderOpenRouter),
			huh.NewOption("Google Cloud", config.ProviderGoogleCloud),
		).Value(&cfg.DefaultProvider),
		huh.NewInput().Title("Default model").Value(&cfg.DefaultModel),
		huh.NewInput().Title("Source language").Value(&cfg.DefaultSourceLanguage),
		huh.NewInput().Title("Target language").Value(&cfg.DefaultTargetLanguage),
		huh.NewConfirm().Title("Enable chunking").Value(&cfg.Chunking.Enabled),
		huh.NewInput().Title("Anthropic API key env var").Value(&anthropic.APIKeyEnv),
		huh.NewInput().Title("Gemini API key env var").Value(&gemini.APIKeyEnv),
		huh.NewInput().Title("OpenRouter API key env var").Value(&openrouter.APIKeyEnv),
		huh.NewInput().Title("OpenRouter app name").Value(&openrouter.AppName),
		huh.NewInput().Title("OpenRouter app URL").Value(&openrouter.AppURL),
		huh.NewInput().Title("Google Cloud credentials env var").Value(&googlecloud.APIKeyEnv),
		huh.NewInput().Title("Google Cloud project ID").Value(&googlecloud.ProjectID),
		huh.NewInput().Title("Google Cloud location").Value(&googlecloud.Location),
		huh.NewInput().Title("Google Cloud mime type").Value(&googlecloud.MimeType),
	))
	if err := form.Run(); err != nil {
		return err
	}
	cfg.Providers[config.ProviderAnthropic] = anthropic
	cfg.Providers[config.ProviderGemini] = gemini
	cfg.Providers[config.ProviderOpenRouter] = openrouter
	cfg.Providers[config.ProviderGoogleCloud] = googlecloud
	return nil
}
