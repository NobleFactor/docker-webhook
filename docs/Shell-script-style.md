<!-- SPDX-FileCopyrightText: 2016-2025 Noble Factor -->
<!-- SPDX-License-Identifier: MIT -->

# Shell Script Style Guide

This document captures the shell scripting conventions used in this repository. Follow these rules when adding or updating scripts under `bin/`, `test/`, or `webhook.config`.

## Purpose

Provide consistent, safe, and maintainable shell scripts that are easy to lint, format, test and install.

## Location

- `bin/` — user-facing helper scripts and installers.
- `test/` — test harnesses and helpers used by `make test`.

## File naming

- Use Verb-Noun style with a hyphen separator and PascalCase words for the noun part, matching existing scripts.

  Examples: `New-DockerNetwork`, `Prepare-WebhookDeployment`, `Test-WebhookReadiness`, `Test-ShellScript`.

- Do not add file extensions (scripts are plain executables).

- Place man pages in `share/man/man1/<script>.1` and completion files in `share/bash-completion/completions/<script>` and `share/zsh/site-functions/_<script>`.

## Header & metadata

- Always start scripts with a POSIX shebang that uses env: `#!/usr/bin/env bash`.

- Include the SPDX license block exactly as other repo scripts show:

```text
########################################################################################################################
# SPDX-FileCopyrightText: 2016-2025 Noble Factor
# SPDX-License-Identifier: MIT
########################################################################################################################
```

- Add a one-line comment describing the script and (when appropriate) a `# shellcheck source=...` annotation for tooling.

## Strict mode and safety

- Use strict options near the top of the script: `set -o errexit -o nounset -o pipefail`.

- Never use `eval` on user-provided input.

- Prefer arrays to whitespace-split strings and always quote expansions: `"$var"`, `"${arr[@]}"`.

## Argument parsing and usage

- Use `bin/Declare-BashScript` for consistent argument parsing and standard `usage`, `note`, `error`, and `success` helpers. Source it like this:

```bash
source "$(dirname "$0")/Declare-BashScript" "$0" \
    "help,foo:,bar:" "h:f:b:" "$@"
```

- Define a `synopsis` variable and call `usage "$synopsis"` for `-h|--help` handling.

- After sourcing, evaluate `set -- "$script_arguments"` and parse options with a `while` + `case` loop.

## Logging and helper functions

- Use the provided helpers:

  - `note "message"` for informational output (writes to stderr like others in repo).
  - `error <rc> "message"` for fatal errors (non-zero rc exits the script).
  - `success "message"` for successful completion messages.

- Keep helper function names lowercase with underscores (e.g., `set_env_from_file`).

## Constants and variables

- Use `declare -r` for constants (e.g., `declare -r script_name`).

- Limit exported variables. Export only when required by downstream tools.

## Exit codes

- Use numeric exit codes consistently:

  - `0` — success
  - `1` — operation failed (generic)
  - `2` — usage/argument parsing errors
  - `3` — resource not found or lookup failure

- Reserve other codes as needed and document them in the script's header or man page.

## Security & secrets

- Do not print secrets or tokens to stdout/stderr.

- Read environment files safely (example in `Prepare-WebhookDeployment`): strip quotes but never `eval` file contents.

## Formatting and Linting

- Format using `shfmt` with 4-space indent: `shfmt -w -i 4 <files>`.

- Lint with `shellcheck -x` and fail CI on issues. Provide a wrapper script (`test/Test-ShellScript`) that finds scripts by shebang and runs `shellcheck -x`.

## Man pages and completions

- Ship manpages in `share/man/man1/<script>.1` and make `usage` prefer `man` if present.

- Provide bash and zsh completions as appropriate in the `share` tree.

## Installation

- Use the `install_script_to_local` helper in `Declare-BashScript` to install scripts and completions into `~/.local` for local testing.

## Testing & artifacts

- Test scripts should write non-destructive artifacts under `test/artifacts/` and leave them for inspection when failures occur.

- Tests that interact with Docker should prefer discovery (find an existing container) over creating/destroying host containers whenever practical.

## Example script skeleton

```bash
#!/usr/bin/env bash
########################################################################################################################
# SPDX-FileCopyrightText: 2016-2025 Noble Factor
# SPDX-License-Identifier: MIT
########################################################################################################################
# My-Script - Brief description
# shellcheck source=../bin/Declare-BashScript

source "$(dirname "$0")/../bin/Declare-BashScript" "$0" "help,option:,flag" "h:o:f:" "$@"

declare -r synopsis="My-Script [--option <val>] [--flag]"
eval set -- "$script_arguments"

while :; do
  case $1 in
    -h|--help)
      usage "$synopsis";;
    --option)
      my_option="$2"; shift 2;;
    --flag)
      my_flag=true; shift 1;;
    --)
      shift; break;;
    *)
      error 2 "Unrecognized option: $1";;
  esac
done

note "Running with option=${my_option:-}" 

# Main logic here

success "Completed"
```

Follow these rules and the repository scripts will remain consistent, lint-clean, and easy to maintain. If you have a special case that requires deviating from these guidelines, document the reason in the script header and link to an issue describing the trade-off.
