#!/usr/bin/env bash

# Naming Policy:
#
# Standalone script — no namespace prefix, no global suffix.
# Functions: cmd.PascalCase (mk.bash convention).
# Locals: camelCase.  Globals: PascalCase.
# _ suffix: contains IFS characters (e.g., Usage_).

Prog=$(basename "$0")   # match what the user called
Version=0.2

read -rd '' Usage_ <<END
Usage:

  $Prog [OPTIONS] [--] COMMAND

  Commands:

    test          -- run all tests

  Options:

    -h | --help     show this message and exit
    -v | --version  show the program version and exit
    -x | --trace    enable debug tracing
END

## commands

cmd.test() {
  mk.Cue go test ./...
}

## boilerplate

source ~/.local/lib/mk.bash 2>/dev/null ||
  eval "$(curl -fsSL https://raw.githubusercontent.com/binaryphile/mk.bash/develop/mk.bash)" ||
  { echo 'fatal: mk.bash not found' >&2; exit 1; }

# enable safe expansion
IFS=$'\n'
set -o noglob

mk.SetProg $Prog
mk.SetUsage "$Usage_"
mk.SetVersion $Version

return 2>/dev/null    # stop if sourced, for interactive debugging
mk.HandleOptions $*   # handle standard options, returns 1-based offset
mk.Main ${*:$?}       # remaining arguments
