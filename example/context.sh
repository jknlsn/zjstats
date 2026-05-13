#!/bin/bash
set -u

PATH="/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"

# Your primary macOS username. Used to detect when you're SSH'd in or
# running as a different user (e.g. root).
# Override by setting ZJSTAT_PREFERRED_USER in your shell rc.
PREFERRED_USER="${ZJSTAT_PREFERRED_USER:-$(id -un)}"

short_hostname() {
  /bin/hostname 2>/dev/null | /usr/bin/perl -pe 's/\..*$//'
}

case "${1:-}" in
  context-ssh)
    if [ -n "${SSH_TTY:-}" ] && [ "$(id -un)" = "$PREFERRED_USER" ]; then
      printf '@%s\n' "$(short_hostname)"
    fi
    ;;
  context-user)
    if [ -z "${SSH_TTY:-}" ] && [ "$(id -un)" != "$PREFERRED_USER" ]; then
      printf '%s@local\n' "$(id -un)"
    fi
    ;;
  context-ssh-user)
    if [ -n "${SSH_TTY:-}" ] && [ "$(id -un)" != "$PREFERRED_USER" ]; then
      printf '%s@%s\n' "$(id -un)" "$(short_hostname)"
    fi
    ;;
  context-local)
    if [ -z "${SSH_TTY:-}" ] && [ "$(id -un)" = "$PREFERRED_USER" ]; then
      printf '@local\n'
    fi
    ;;
  *)
    printf 'usage: %s {context-ssh|context-user|context-ssh-user|context-local}\n' "$0" >&2
    exit 1
    ;;
esac
