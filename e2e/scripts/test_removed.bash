#!/usr/bin/env bash
set -eu

: "${VENOM_VAR_binpath:? Please set VENOM_VAR_binpath to updatecli binary dirname}"
: "${VENOM_VAR_rootpath:=../..}"

cd "$VENOM_VAR_rootpath"

for command in apply diff prepare show; do
  if output=$("$VENOM_VAR_binpath/updatecli" "$command" 2>&1); then
    printf 'Removed command unexpectedly succeeded: %s\n' "$command"
    exit 1
  fi
  [[ "$output" == *"unknown command"* ]]
done

for manifest in githubPullrequest.yaml json.yaml transformers.yaml; do
  if output=$("$VENOM_VAR_binpath/updatecli" --disable-version-check pipeline diff \
    --config "e2e/updatecli.d/removed.d/$manifest" 2>&1); then
    printf 'Removed manifest setting unexpectedly succeeded: %s\n' "$manifest"
    exit 1
  fi
  [[ "$output" == *"removed in v1"* ]]
done

printf 'Removed interfaces rejected\n'
