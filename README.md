# nabr

A dynamic CLI tool that turns YAML-defined API endpoints into shell commands. Define your HTTP APIs once in a config file, then call them as CLI subcommands with flags.

## Install

Requires Go 1.25+.

```bash
go build -o nabr .
# optionally move to your PATH
mv nabr /usr/local/bin/
```

## Quick Start

Create a config file at `~/.config/nabr/config.yaml`:

```yaml
commands:
  - name: get-ip
    description: Get your public IP address
    method: GET
    url: https://httpbin.org/ip
```

Run it:

```bash
nabr get-ip
# HTTP 200
# {
#   "origin": "203.0.113.42"
# }
```

## Config File

The default config path is `~/.config/nabr/config.yaml`. Override it with `--config`:

```bash
nabr --config ./my-apis.yaml get-ip
```

Each command in the YAML has these fields:

| Field | Required | Description |
|---|---|---|
| `name` | yes | The CLI subcommand name |
| `description` | no | Short help text shown in `--help` |
| `method` | yes | HTTP method (`GET`, `POST`, `PUT`, `DELETE`, etc.) |
| `url` | yes | Endpoint URL, supports `{param}` path parameters |
| `headers` | no | Default headers as `key: value` pairs |
| `body` | no | Default request body string |
| `query_params` | no | Default query parameters as `key: value` pairs |

### Example config

```yaml
commands:
  - name: get-user
    description: Fetch a user by ID
    method: GET
    url: https://api.example.com/users/{id}
    headers:
      Authorization: Bearer my-token
    query_params:
      verbose: "true"

  - name: create-post
    description: Create a new blog post
    method: POST
    url: https://api.example.com/users/{userId}/posts
    headers:
      Content-Type: application/json
    body: '{"title": "New Post", "draft": true}'
```

## Usage

### Path Parameters

URL placeholders like `{id}` automatically become required flags:

```bash
nabr get-user --id 42
# GET https://api.example.com/users/42
```

Multiple path params each get their own flag:

```bash
nabr create-post --userId 42
# POST https://api.example.com/users/42/posts
```

### Query Parameters (`-q`)

Add or override query parameters at runtime. Repeatable.

```bash
nabr get-user --id 42 -q verbose=false -q fields=name,email
```

Config-defined query params are merged with CLI flags. CLI flags override config values for the same key.

### Headers (`-H`)

Add or override headers at runtime. Repeatable.

```bash
nabr get-user --id 42 -H "Authorization=Bearer new-token" -H "Accept=application/json"
```

### Body (`-b`)

Override the request body at runtime:

```bash
nabr create-post --userId 42 -b '{"title": "Updated Title", "draft": false}'
```

### Raw Output (`--raw`)

By default, JSON responses are pretty-printed. Use `--raw` to get the unformatted response:

```bash
nabr get-ip --raw
```

## Global Flags

| Flag | Description |
|---|---|
| `--config <path>` | Path to config file (default: `~/.config/nabr/config.yaml`) |
| `--raw` | Output raw response without pretty-printing |
| `-h`, `--help` | Help for nabr or any subcommand |

## Shell Completion

nabr supports shell completion via Cobra:

```bash
# Bash
nabr completion bash > /etc/bash_completion.d/nabr

# Zsh
nabr completion zsh > "${fpath[1]}/_nabr"

# Fish
nabr completion fish > ~/.config/fish/completions/nabr.fish
```

## Development

```bash
go build ./...     # build
go test ./...      # run tests
go vet ./...       # lint
```
