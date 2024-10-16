# direnv-tiny

direnv-tiny is a minimal implementation inspired by [direnv](https://github.com/direnv/direnv).

direnv-tiny was developed under the hypothesis that a minimal codebase could be achieved by focusing solely on two core mechanisms: exporting environment variables when a .envrc file is present in the current directory, and unsetting those variables when moving to a directory without a .envrc file.

## Installation

Build for tool:

```
$ go build
```

## Usage

1. Add the following to your shell configuration file (e.g., `.bashrc`, `.zshrc`):
   ```bash
   eval "$(direnv-tiny hook)"
   ```

2. Create a `.envrc` file in any directory where you want to set specific environment variables:
   ```
   export PROJECT_ROOT=$(pwd)
   export DATABASE_URL=postgresql://localhost/mydb
   ```

3. Navigate to the directory, and direnv-tiny will automatically load the environment variables.

## Commands

- `direnv-tiny hook`: Outputs the shell hook that needs to be evaluated.
- `direnv-tiny export`: Exports the environment variables based on the current directory.

## Debug Mode

To enable debug logging, set the `DIRENV_TINY_DEBUG` environment variable to `1`:

```bash
export DIRENV_TINY_DEBUG=1
```

## License

This project is licensed under the MIT License - see the [LICENSE](https://opensource.org/license/mit) for details.
