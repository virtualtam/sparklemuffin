# Project Structure
## Overview
Source code is broken down into several packages:

```shell
.
├── cmd       # Command-line application, HTTP servers, Web application
├── internal  # Private packages and test helpers
└── pkg       # Domain packages
```

## `cmd` - Command-line application
```shell
cmd
└── sparklemuffin
    ├── command    # Command-line application commands and sub-commands (CLI parser)
    ├── config     # Configuration utilities
    ├── http       # HTTP servers: metrics, Web application
    ├── main.go    # Command-line entrypoint
    └── version    # Version detection utilities
```

## `internal` - Application-specific and private packages
```shell
internal
├── hash            # Cryptographically secure hash helpers
├── paginate        # Pagination utilities
├── rand            # Cryptographically secure pseudo-random helpers
└── repository
│   └── postgresql  # PostgreSQL database persistence layer (repository)
└── test            # Helpers for unitary and integration tests
```


## `pkg` - Domain packages
```shell
pkg
├── bookmark  # Web bookmark management
├── feed      # Feed subscription management
├── session   # User session persistence
└── user      # User and permission management
```
