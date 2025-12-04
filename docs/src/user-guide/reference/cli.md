# Command-line flags
SparkleMuffin is provided as a single-binary command-line application,
that provides commands to:

- run the Web server
- create administrator users
- apply database migrations
- display information about how the program was built
- etc.

To see which commands and global flags are available, run the `sparklemuffin --help`
command:

```shell
$ sparklemuffin --help

SparkleMuffin - Web Bookmark Manager

Usage:
  sparklemuffin [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  createadmin Create a user with administration privileges
  help        Help about any command
  migrate     Initialize database and run migrations
  run         Start the HTTP server
  version     Display the prorgam version

Flags:
      --db-addr string       Database address (host:port) (default "localhost:15432")
      --db-name string       Database name (default "sparklemuffin")
      --db-password string   Database password (default "sparklemuffin")
      --db-sslmode string    Database sslmode (default "disable")
      --db-user string       Database user (default "sparklemuffin")
  -h, --help                 help for sparklemuffin
      --hmac-key string      Secret key for HMAC session token hashing (default "hmac-secret-key")
      --log-level string     Log level (trace, debug, info, warn, error, fatal, panic) (default "info")

Use "sparklemuffin [command] --help" for more information about a command.
```

To get more information about a specific command, run `sparklemuffin <command> --help`:

```shell
$ sparklemuffin run --help

Start the HTTP server

Usage:
  sparklemuffin run [flags]

Flags:
  -h, --help                         help for run
      --listen-addr string           Listen to this address (host:port) (default "0.0.0.0:8080")
      --metrics-listen-addr string   Listen to this address for Prometheus metrics (host:port) (default "127.0.0.1:8081")
      --public-addr string           Public HTTP address (if behind a proxy) (default "http://localhost:8080")

Global Flags:
      --db-addr string       Database address (host:port) (default "localhost:15432")
      --db-name string       Database name (default "sparklemuffin")
      --db-password string   Database password (default "sparklemuffin")
      --db-sslmode string    Database sslmode (default "disable")
      --db-user string       Database user (default "sparklemuffin")
      --hmac-key string      Secret key for HMAC session token hashing (default "hmac-secret-key")
      --log-level string     Log level (trace, debug, info, warn, error, fatal, panic) (default "info")
```
