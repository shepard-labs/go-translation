# aitranslate

`aitranslate` is a local Go CLI for AI-based translation using [`github.com/shepard-labs/go-ai-sdk`](https://github.com/shepard-labs/go-ai-sdk).

It translates strings, files, and piped stdin through Anthropic, Gemini, or OpenRouter. The tool is local-first: it has no hosted service, no daemon, no telemetry, and no cloud persistence. Your config stays on your machine, and translation requests are sent only to the provider you choose.

## Features

- Translate direct strings, files, or stdin
- Use Anthropic, Gemini, or OpenRouter
- Configure provider, model, languages, token limits, timeout, and chunking
- Store local config in `~/.config/aitranslate/config.yaml`
- Initialize config with a guided terminal form
- Open an interactive Bubble Tea TUI with `aitranslate tui`
- Preserve Markdown, code blocks, frontmatter, placeholders, variables, URLs, lists, and tables where possible
- Split large files into deterministic chunks and reassemble translated output
- Refuse to overwrite output files unless `--force` is passed
- Avoid printing secret values in config output

## Install

Build the binary from this repository:

```bash
go build -o aitranslate ./cmd/aitranslate
```

Optional local install:

```bash
go install ./cmd/aitranslate
```

Verify the CLI:

```bash
./aitranslate --help
```

## Quick Start

Set at least one provider API key:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

Translate a string:

```bash
./aitranslate "Hello world" --to Spanish
```

Translate explicit text:

```bash
./aitranslate --text "Hello world" --to Spanish
```

Translate a file to stdout:

```bash
./aitranslate --file ./input.md --to French
```

Translate a file and save the output:

```bash
./aitranslate --file ./input.md --to French --out ./input.fr.md
```

Translate piped stdin:

```bash
cat input.md | ./aitranslate --to Japanese
```

## Providers

Supported providers:

```text
anthropic
gemini
openrouter
googlecloud
```

List providers:

```bash
./aitranslate providers list
```

Show example model names:

```bash
./aitranslate models examples
```

Provider examples:

```bash
./aitranslate "Hello world" --to Spanish --provider anthropic --model claude-sonnet-4-6
./aitranslate "Hello world" --to Spanish --provider gemini --model gemini-3.1-pro
./aitranslate "Hello world" --to Spanish --provider openrouter --model anthropic/claude-sonnet-4.6
```

## Provider Setup

Anthropic:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
./aitranslate "Good morning" --to German --provider anthropic
```

Gemini:

```bash
export GEMINI_API_KEY=...
./aitranslate "Good morning" --to German --provider gemini --model gemini-3.1-pro
```

OpenRouter:

```bash
export OPENROUTER_API_KEY=...
./aitranslate "Good morning" --to German --provider openrouter --model anthropic/claude-sonnet-4.6
```

Google Cloud (Cloud Translation API v3):

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json  # or rely on ADC
./aitranslate "Good morning" --to German --provider googlecloud
```

`googlecloud` requires `providers.googlecloud.project_id` in config and accepts `--to` values as ISO codes (`de`) or English names (`German`).

## Configuration

Create a local config with guided prompts:

```bash
./aitranslate config init
```

If a config file already exists, `config init` refuses to overwrite it unless `--force` is used:

```bash
./aitranslate --force config init
```

Default config path:

```text
~/.config/aitranslate/config.yaml
```

Use a custom config path:

```bash
./aitranslate --config ./aitranslate.yaml "Hello" --to Spanish
```

Show the resolved config without printing secret values:

```bash
./aitranslate config show
```

Example config:

```yaml
default_provider: anthropic
default_model: claude-sonnet-4-6
default_source_language: auto
default_target_language: English
timeout: 120s
max_output_tokens: 4096
temperature: 0

chunking:
  enabled: true
  max_chars: 12000

providers:
  anthropic:
    api_key_env: ANTHROPIC_API_KEY
    default_model: claude-sonnet-4-6

  gemini:
    api_key_env: GEMINI_API_KEY
    default_model: gemini-3.1-pro

  openrouter:
    api_key_env: OPENROUTER_API_KEY
    default_model: anthropic/claude-sonnet-4.6
    app_name: aitranslate
    app_url: ""
```

Configuration precedence:

1. CLI flags
2. Environment variables
3. Config file
4. Built-in defaults

## Environment Variables

`aitranslate` supports these app-level environment variables:

```bash
AITRANSLATE_PROVIDER=anthropic
AITRANSLATE_MODEL=claude-sonnet-4-6
AITRANSLATE_FROM=auto
AITRANSLATE_TO=Spanish
AITRANSLATE_CONFIG=/custom/config.yaml
```

Provider credentials:

```bash
ANTHROPIC_API_KEY=...
GEMINI_API_KEY=...
OPENROUTER_API_KEY=...
```

The config stores the credential environment variable names, not the raw secret values.

## Commands

Root translation command:

```bash
aitranslate [text] [flags]
```

Subcommands:

```bash
aitranslate config init
aitranslate config show
aitranslate providers list
aitranslate models examples
aitranslate tui
```

Common flags:

```text
--provider string           Provider to use: anthropic, gemini, openrouter, or googlecloud
--model string              Model to use
--from string               Source language, or auto
--to string                 Target language
--config string             Config file path
--timeout duration          Request timeout
--max-output-tokens int     Maximum output tokens
--temperature float         Sampling temperature
--text string               Text to translate
--file string               File to translate
--out string                Output file path
--force                     Overwrite existing output/config files
--verbose                   Print extra status information
```

Input validation rules:

- Use exactly one input source: positional text, `--text`, `--file`, or piped stdin.
- Pass `--to` or set `default_target_language` in config.
- Set the provider API key env var for the selected provider.
- Existing output files are not overwritten unless `--force` is passed.

## TUI

Open the interactive TUI:

```bash
./aitranslate tui
```

The TUI lets you edit provider, model, source language, target language, and input text, then run translation from the terminal UI.

Keybindings:

```text
tab          next field
shift+tab    previous field
ctrl+t       translate
q            quit
ctrl+c       quit
```

## Large Files

Chunking is enabled by default:

```yaml
chunking:
  enabled: true
  max_chars: 12000
```

For inputs larger than `max_chars`, `aitranslate` splits by paragraph boundaries first, then line boundaries, then hard character boundaries. Chunks are translated sequentially and reassembled in order.

## Exit Codes

```text
0 success
1 provider/runtime error
2 invalid CLI usage/config
3 input/output file error
4 missing credentials
```

## Troubleshooting

Missing credentials:

```text
missing credentials: set ANTHROPIC_API_KEY or configure providers.anthropic.api_key_env
```

Fix by exporting the expected provider key or changing the provider config.

Missing target language:

```text
missing target language: pass --to or set default_target_language
```

Fix by passing `--to Spanish` or setting `default_target_language` in the config file.

Input conflict:

```text
input conflict: use only one of positional text, --text, --file, or stdin
```

Fix by choosing one input mode.

Output file exists:

```text
output file exists: pass --force to overwrite ./input.es.md
```

Fix by choosing a new output path or passing `--force`.

## Development

Run checks:

```bash
gofmt -w .
go test ./...
go build -o aitranslate ./cmd/aitranslate
```

Optional provider integration tests should be kept behind explicit environment flags so the normal test suite does not require real API keys.
