# Project Structure
## Overview
The Go source code is broken down into several top-level packages:

```shell
.
├── cmd       # Command-line application
├── internal  # Private packages and test helpers
└── pkg       # Domain packages
```

## `cmd` - Command-line application
```shell
cmd
└── sparklemuffin
    ├── command    # Command-line application commands and sub-commands (CLI parser)
    ├── config     # Configuration utilities
    └── main.go    # Command-line entrypoint
```

## `internal` - Application-specific and private packages
```shell
internal
├── hash            # Cryptographically secure hash helpers
├── http            # HTTP servers: monitoring, Web application
├── paginate        # Pagination utilities
├── rand            # Cryptographically secure pseudo-random helpers
├── repository
│   └── postgresql  # PostgreSQL database persistence layer (repository)
├── test            # Helpers for unitary and integration tests
├── textkit         # Text processing utilities
└── version         # Version detection utilities
```


## `pkg` - Domain packages
```shell
pkg
├── bookmark  # Web bookmark management
├── feed      # Feed subscription management
├── session   # User session persistence
└── user      # User and permission management
```
