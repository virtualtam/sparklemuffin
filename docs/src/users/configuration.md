# Configuration
## Configuration formats
SparkleMuffin can be configured by:

- using a configuration file;
- setting environment variables, e.g. `SPARKLEMUFFIN_LOG_LEVEL=debug`;
- setting POSIX flags, e.g. `--log-level debug`.

### Configuration variable precedence
Configuration variables are evaluated in this order:

- program defaults (lowest precedence);
- configuration file;
- command-line flags;
- environment variables (highest precedence).

### Naming convention for configuration variables
All configuration variables are specified as program flags (see [command-line flags](./cli.md)),
from which environment variables names and configuration file keys are derived:

| Command-line flag   | Environment Variable            | Configuration File |
| ------------------- | ------------------------------- | ------------------ |
| `--example`         | `SPARKLEMUFFIN_EXAMPLE`         | `example: true`    |
| `--log-level debug` | `SPARKLEMUFFIN_LOG_LEVEL=debug` | `log-level: debug` |

## Configuration file
- TODO: add CLI flag to specify a configuration file
- TODO: add CLI command to generate a configuration file with default values
- TODO: specify configuration file format (TOML?)
- TODO: add commented configuration file to SCM
- TODO: add commented configuration file to docs (section on this page)
