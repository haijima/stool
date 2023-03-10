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

or you can download binaries from [Releases](https://github.com/haijima/stool/releases).

## Commands and Options

``` sh
stool [command]
```

### Commands

- `scenario`: Show the access patterns of users
- `transition`: Show the transition between endpoints
- `trend`: Show the count of accesses for each endpoint over time

### Options

#### Global Options

- `--config string` : Config file (default is $HOME/.stool.yaml)
- `-f, --file string` : Access log file to profile.
- `--ignore_patterns strings` : Comma-separated list of regular expression patterns to ignore URIs
- `-m, --matching_groups strings` : Comma-separated list of regular expression patterns to group matched URIs. For
  example: `--matching_groups "/users/.*,/items/.*"`.
  `--time_format string` : The format to parse time field on log file (default `02/Jan/2006:15:04:05 -0700`).
  Options for scenario

#### Options for `stool scenario`

- `--format` : The output format (dot, csv) (default "dot").
  Options for transition

#### Options for `stool transition`

- `--format` : The output format (dot, csv) (default "dot").
  Options for trend

#### Options for `stool trend`

- `-i, --interval` : The time (in seconds) of the interval. Access counts are cumulated at each interval. (default `5`).

### Configuration and Customization

Instructions on how to configure and customize the tool can be found in the `config.md` file.

## Examples

``` sh
stool scenario --file path/to/access.log --matching_groups "/users/.*,/items/.*" --format dot | dot -T svg -o scenario.svg && open scenario.svg

stool transition --file path/to/access.log --matching_groups "/users/.*,/items/.*" --format dot | dot -T svg -o transition.svg && open transition.svg

stool trend --file path/to/access.log --matching_groups "/users/.*,/items/.*" --interval 
```

## Prerequisites

### Graphviz

[Graphviz](https://graphviz.org/) is required to visualize `.dot` files that are generated by `stool scenario`
or `stool transition` command.

### LTSV Log format

`stool` can handle LTSV formatted log file. The example of Nginx log setting is shown below:

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
