# stool

[![CI Status](https://github.com/haijima/stool/workflows/CI/badge.svg?branch=main)](https://github.com/haijima/stool/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/haijima/stool.svg)](https://pkg.go.dev/github.com/haijima/stool)
[![Go report](https://goreportcard.com/badge/github.com/haijima/stool)](https://goreportcard.com/report/github.com/haijima/stool)

stool is a CLI tool to analyze access pattern from Nginx access log.

## Installation

```
go install github.com/haijima/stool@latest
```

or you can download binaries from [Releases](https://github.com/haijima/stool/releases).

## Commands and Options

```
stool [command]
```

### Commands

- `scenario`: Show the access patterns of users
- `transition`: Show the transition between endpoints
- `trend`: Show the count of accesses for each endpoint over time

## License

This tool is licensed under the MIT License. See the `LICENSE` file for details.
