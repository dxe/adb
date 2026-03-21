# Load shared shell functions
[ -f "$HOME/.bash_adb_functions" ] && source "$HOME/.bash_adb_functions"

# Add git support for bash-completion. Assumes bash-completion apt package is
# installed. devcontainer.json should install it automatically.
source /usr/share/bash-completion/completions/git

### Useful aliases for development ###

# Run ADB CLI
alias adb="go run /workspace/cli"

# Load ADB bash completion for the adb alias in interactive shells.
if [[ $- == *i* ]]; then
  # Store the generated completion script in a persistent cache so new shells
  # can source a file directly instead of running `go run` every time.
  _adb_completion_cache="${XDG_CACHE_HOME:-$HOME/.cache}/adb-completion.bash"
  _adb_completion_needs_refresh=0

  if [[ -d /workspace/cli ]]; then
    # Refresh if the cache is missing/empty, or if any Go source/module file in
    # the CLI project is newer than the cached completion script.
    if [[ ! -s "$_adb_completion_cache" ]]; then
      _adb_completion_needs_refresh=1
    elif find /workspace/cli -type f \( -name '*.go' -o -name 'go.mod' -o -name 'go.sum' \) -newer "$_adb_completion_cache" -print -quit 2>/dev/null | grep -q .; then
      _adb_completion_needs_refresh=1
    fi

    # Only try to regenerate when Go is installed. Write to a temporary file
    # first so a failed generation does not leave a partial cache behind.
    if [[ $_adb_completion_needs_refresh -eq 1 ]] && command -v go >/dev/null 2>&1; then
      mkdir -p "$(dirname "$_adb_completion_cache")"
      if go run /workspace/cli completion bash >"${_adb_completion_cache}.tmp" 2>/dev/null; then
        mv "${_adb_completion_cache}.tmp" "$_adb_completion_cache"
      fi
      rm -f "${_adb_completion_cache}.tmp"
    fi
  fi

  # If a cached completion file exists, load it even when regeneration was
  # skipped (for example because Go is not currently installed).
  [[ -s "$_adb_completion_cache" ]] && source "$_adb_completion_cache"
  unset _adb_completion_cache _adb_completion_needs_refresh
fi
