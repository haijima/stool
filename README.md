# stool

[![CI Status](https://github.com/haijima/stool/workflows/CI/badge.svg?branch=main)](https://github.com/haijima/stool/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/haijima/stool.svg)](https://pkg.go.dev/github.com/haijima/stool)
[![Go report](https://goreportcard.com/badge/github.com/haijima/stool)](https://goreportcard.com/report/github.com/haijima/stool)

`stool` is a CLI tool to analyze access pattern from Nginx access log.

## Installation

You can install stool using the following command:

``` sh
go install github.com/haijima/stool@latest
```

MacOS users can install stool using [Homebrew](https://brew.sh/) (See also [haijima/homebrew-tap](http://github.com/haijima/homebrew-tap)):

``` sh
brew install haijima/tap/stool
```

or you can download binaries from [Releases](https://github.com/haijima/stool/releases).

## Commands and Options

### Commands

- `stool param`: Show the parameter statistics for each endpoint
- `stool scenario`: Show the access patterns of users
- `stool transition`: Show the transition between endpoints
- `stool trend`: Show the count of accesses for each endpoint over time
- `stool genconf`: Generate configuration file

### Options

#### Global Options

- `--config string` : Config file (default is `$XDG_CONFIG_HOME/.stool.yaml`)
- `--no_color`: Disable colorized output
- `-q, --quiet`: Quiet output
- `--verbosity int`: Verbosity level (default `0`)

#### Options for `stool genconf`

- `--capture-group-name` : Add names to captured groups like `"(?P<name>pattern)"` (default `false`)
- `--format string` : The output format {`toml`|`yaml`|`json`|`flag`} (default `"yaml"`)

#### Options for `stool param`

- `-f, --file string` : Access log file to profile.
- `--filter string` : Filter log lines by regular expression
- `--format string`: The stat output format {`table`|`md`|`csv`|`tsv`} (default `"table"`)
- `--log_labels stringToString` : Comma-separated list of key=value pairs to override log labels (default `[]`)
- `-m, --matching_groups strings` : Comma-separated list of regular expression patterns to group matched URIs. For
- `-n, --num int`: The number of parameters to show (default `5`)
- `--stat`: Show statistics of the parameters
- `--time_format string` : The format to parse time field on log file (default `"02/Jan/2006:15:04:05 -0700"`).
- `-t, --type string`: The type of the parameter {`path`|`query`|`all`} (default `"all"`)
  example: `--matching_groups "/users/.*,/items/.*"`.

#### Options for `stool scenario`

- `-f, --file string` : Access log file to profile.
- `--filter string` : Filter log lines by regular expression
- `--format string` : The output format {`dot`|`mermaid`|`csv`} (default `"dot"`).
- `--log_labels stringToString` : Comma-separated list of key=value pairs to override log labels (default `[]`)
- `-m, --matching_groups strings` : Comma-separated list of regular expression patterns to group matched URIs. For
  example: `--matching_groups "/users/.*,/items/.*"`.
- `--palette` : Use color palette for each endpoint (default `false`)
- `--time_format string` : The format to parse time field on log file (default `"02/Jan/2006:15:04:05 -0700"`).

#### Options for `stool transition`

- `-f, --file string` : Access log file to profile.
- `--filter string` : Filter log lines by regular expression
- `--format string` : The output format {`dot`|`mermaid`|`csv`} (default `"dot"`).
- `--log_labels stringToString` : Comma-separated list of key=value pairs to override log labels (default `[]`)
- `-m, --matching_groups strings` : Comma-separated list of regular expression patterns to group matched URIs. For
  example: `--matching_groups "/users/.*,/items/.*"`.
- `--time_format string` : The format to parse time field on log file (default `"02/Jan/2006:15:04:05 -0700"`).

#### Options for `stool trend`

- `-f, --file string` : Access log file to profile.
- `--filter string` : Filter log lines by regular expression
- `--format string` : The output format {`table`|`md`|`csv`} (default `"table"`)
- `-i, --interval int` : The time (in seconds) of the interval. Access counts are cumulated at each interval. (
  default `5`).
- `--log_labels stringToString` : Comma-separated list of key=value pairs to override log labels (default `[]`)
- `-m, --matching_groups strings` : Comma-separated list of regular expression patterns to group matched URIs. For
- `--sort string` : Comma-separated list of `"<sort keys>:<order>"` Sort keys
  are {`method`|`uri`|`sum`|`count0`|`count1`|`countN`}. Orders are [`asc`|`desc`]. e.g. `"sum:desc,count0:asc"` (
  default `"sum:desc"`)
  example: `--matching_groups "/users/.*,/items/.*"`.
- `--time_format string` : The format to parse time field on log file (default `"02/Jan/2006:15:04:05 -0700"`).

### Config file

If `--config` option is specified, **only** the passed file is used as the default options.

Otherwise, you can use a global config file and/or a project specific config file.
They are merged and used as the default options.
If the same option is specified in both config files, the option in the project specific config file is used.

The global config file is searched in the following order:

1. `$XDG_CONFIG_HOME/stool/config.yaml`
2. `$HOME/.config/stool/config.yaml` when `$XDG_CONFIG_HOME` is not set
3. `$HOME/.stool.yaml`

The project specific config file is `$CURRENT_DIR/.stool.yaml`.

Not only [YAML](https://yaml.org/) but also [JSON](https://www.json.org/json-en.html) and [TOML](https://toml.io/en/) format are supported.

#### Example

When the following config files exist and `--config` option is not specified
```yaml
# $XDG_CONFIG_HOME/stool/config.yaml
optionA: 1
optionB: 1

# $HOME/.stool.yaml
optionA: 2
optionB: 2

# $CURRENT_DIR/.stool.yaml
optionB: 3
optionC: 3
```

The merged config is as follows
```yaml
optionA: 1 # from $XDG_CONFIG_HOME/stool/config.yaml (prior to $HOME/.stool.yaml as global config) 
optionB: 3 # overridden by $CURRENT_DIR/.stool.yaml as project specific config
optionC: 3 # from $CURRENT_DIR/.stool.yaml as project specific config
```


## Examples

``` sh
stool param --file path/to/access.log --num 10 --stat "/users/(?P<userId>[^/]+)$"

stool scenario --file path/to/access.log --matching_groups "/users/.*,/items/.*" --format dot | dot -T svg -o scenario.svg && open scenario.svg

stool transition --file path/to/access.log --matching_groups "/users/.*,/items/.*" --format dot | dot -T svg -o transition.svg && open transition.svg

stool trend --file path/to/access.log --matching_groups "/users/.*,/items/.*" --interval 10

stool genconf path/to/main.go --format yaml >> .stool.yaml
```

## Prerequisites

### Graphviz

[Graphviz](https://graphviz.org/) is required to visualize `.dot` files that are generated by `stool scenario`
or `stool transition` command.

### Nginx Log format

1. `ngx_http_userid_module` is required to get the user ID from the cookie.
   See [Module ngx\_http\_userid\_module](http://nginx.org/en/docs/http/ngx_http_userid_module.html) for details.
2. `stool` can handle [LTSV](http://ltsv.org/) formatted log file.

The example of Nginx log setting is shown below:

```nginx configuration
userid on; # Enable $uid_got and $uid_set

log_format ltsv "time:$time_local"
                "\thost:$remote_addr"
                "\tforwardedfor:$http_x_forwarded_for"
                "\treq:$request"
                "\tstatus:$status"
                "\tmethod:$request_method"
                "\turi:$request_uri"
                "\tsize:$body_bytes_sent"
                "\treferer:$http_referer"
                "\tua:$http_user_agent"
                "\treqtime:$request_time"
                "\tcache:$upstream_http_x_cache"
                "\truntime:$upstream_http_x_runtime"
                "\tapptime:$upstream_response_time"
                "\tuidgot:$uid_got"
                "\tuidset:$uid_set"
                "\tvhost:$host";
```

## License

This tool is licensed under the MIT License. See the `LICENSE` file for details.
