#!/usr/bin/env bash

# Naming Policy:
#
# Standalone script — no namespace prefix, no global suffix.
# Functions: cmd.PascalCase (mk.bash convention).
# Locals: camelCase.  Globals: PascalCase.
# _ suffix: contains IFS characters (e.g., Usage_).

Prog=$(basename "$0")   # match what the user called
Version=0.1

read -rd '' Usage_ <<END
Usage:

  $Prog [OPTIONS] [--] COMMAND

  Commands:

    task          -- publish a task event
    done          -- publish a task-done event
    open          -- list open tasks
    audit         -- full task reconciliation
    claim         -- claim a task
    claims        -- list active claims

  Options:

    -h | --help     show this message and exit
    -v | --version  show the program version and exit
    -x | --trace    enable debug tracing
END

BinDir=bin
ProjectRoot=$(cd "$(dirname "$0")" && pwd)
TaskStream=tasks.$(basename "$ProjectRoot")
TaskAudit=$ProjectRoot/$BinDir/task-audit

## task stream commands

cmd.task() {
  [[ $# -ge 1 ]] || { echo "usage: $Prog task <description>" >&2; return 1; }
  era publish -s $TaskStream --type task "$1"
}

cmd.done() {
  [[ $# -ge 1 ]] || { echo "usage: $Prog done <id>[,<id>...] [evidence]" >&2; return 1; }
  local ids=$1
  local evidence=${2:-}
  # validate: all comma-separated parts are numeric
  local parts
  IFS=, read -ra parts <<< "$ids"
  local id
  for id in "${parts[@]}"; do
    [[ $id =~ ^[0-9]+$ ]] || { echo "error: '$id' is not a numeric ID" >&2; return 1; }
  done
  # build payload with #id references
  local refs
  refs=$(printf '%s' "$ids" | sed 's/[0-9]\{1,\}/#&/g')
  local payload="Task $refs — done."
  [[ -n $evidence ]] && payload="Task $refs — done. $evidence"
  # era splits -m on commas, so use + as separator in refs value
  local refsVal=${ids//,/+}
  era publish -s $TaskStream --type task-done -m "refs=$refsVal" "$payload"
}

cmd.open() {
  $TaskAudit open $TaskStream
}

cmd.audit() {
  $TaskAudit audit $TaskStream
}

cmd.claim() {
  [[ $# -ge 2 ]] || { echo "usage: $Prog claim <id> <name>" >&2; return 1; }
  [[ $1 =~ ^[0-9]+$ ]] || { echo "error: '$1' is not a numeric ID" >&2; return 1; }
  if ! $TaskAudit check $TaskStream "$1" 2>/dev/null; then
    echo "warning: task #$1 already claimed" >&2
  fi
  era publish -s $TaskStream --type claim -m "refs=$1,claimer=$2" "Claimed #$1 by $2"
}

cmd.claims() {
  $TaskAudit claims $TaskStream
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
