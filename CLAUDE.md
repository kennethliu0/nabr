# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is nabr?

A dynamic CLI tool that reads API command definitions from a YAML config file (`~/.config/nabr/config.yaml` by default) and registers each as a Cobra subcommand. Users define HTTP endpoints declaratively and invoke them as CLI commands with path parameter flags.

## Build & Run

```bash
go build -o nabr .      # build the binary
go run .                 # run without building
go test ./...            # run all tests
go vet ./...             # lint
```

## Architecture

- **`main.go`** — Entrypoint, calls `cmd.Execute()`.
- **`cmd/`** — Cobra command setup. `root.go` loads config at init time, dynamically registers one subcommand per YAML-defined API command. Path parameters (`{param}` in URLs) become required `--param` flags.
- **`config/`** — YAML config loading via Viper. `Config` holds a list of `Command` structs (name, description, method, url, headers, body, query_params).
- **`request/`** — HTTP execution. Extracts path params from URL templates, substitutes values, executes the request, and formats JSON output. `--raw` flag skips pretty-printing.

## Config file format

Commands are defined in YAML with fields: `name`, `description`, `method`, `url`, `headers`, `body`, `query_params`. URL path parameters use `{paramName}` syntax.

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- `github.com/spf13/viper` — Config file parsing
- Go 1.25+
