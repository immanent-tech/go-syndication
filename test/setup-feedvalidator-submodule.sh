#!/usr/bin/env bash
#
# setup-feedvalidator-submodule.sh
#
# Adds (or re-initializes) the w3c/feedvalidator repo as a git submodule,
# sparse-checked-out so only the testcases/ directory is populated on disk.
#
# Usage:
#   ./setup-feedvalidator-submodule.sh [path]
#
#   path   Optional. Where to place the submodule.
#          Defaults to "third_party/feedvalidator".
#
# Safe to re-run: if the submodule is already registered, this script
# will just (re)apply the sparse-checkout settings and check it out.

set -euo pipefail

REPO_URL="https://github.com/w3c/feedvalidator.git"
SUBMODULE_PATH="${1:-test/feedvalidator}"
SPARSE_DIR="testcases"

# --- sanity checks -----------------------------------------------------

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "Error: this script must be run from inside a git repository." >&2
  exit 1
fi

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

# --- add or re-use the submodule ---------------------------------------

if git config --file .gitmodules --get "submodule.${SUBMODULE_PATH}.url" >/dev/null 2>&1; then
  echo "Submodule already registered at '${SUBMODULE_PATH}'. Ensuring it's initialized..."
  git submodule update --init --no-fetch "${SUBMODULE_PATH}" 2>/dev/null || \
    git submodule update --init "${SUBMODULE_PATH}"
else
  echo "Adding submodule '${REPO_URL}' at '${SUBMODULE_PATH}'..."
  git submodule add --depth 1 "${REPO_URL}" "${SUBMODULE_PATH}"
fi

# --- configure sparse-checkout inside the submodule ---------------------

echo "Configuring sparse-checkout for '${SPARSE_DIR}/' only..."
git -C "${SUBMODULE_PATH}" sparse-checkout init --cone
git -C "${SUBMODULE_PATH}" sparse-checkout set "${SPARSE_DIR}"

# --- checkout the default branch ----------------------------------------

DEFAULT_BRANCH="$(git -C "${SUBMODULE_PATH}" remote show origin 2>/dev/null \
  | sed -n '/HEAD branch/s/.*: //p')"
DEFAULT_BRANCH="${DEFAULT_BRANCH:-main}"

echo "Checking out '${DEFAULT_BRANCH}'..."
git -C "${SUBMODULE_PATH}" checkout "${DEFAULT_BRANCH}"

# --- stage the submodule reference in the parent repo --------------------

git add .gitmodules "${SUBMODULE_PATH}"

echo
echo "Done. '${SUBMODULE_PATH}/${SPARSE_DIR}' is populated on disk."
echo "Review the staged changes and commit, e.g.:"
echo "  git commit -m \"Add feedvalidator submodule, sparse-checked-out to ${SPARSE_DIR}\""
echo
echo "Note: sparse-checkout settings are local to this clone (stored in"
echo "${SUBMODULE_PATH}/.git/info/sparse-checkout). Anyone else who clones"
echo "this repo should run this script again after 'git submodule update --init'"
echo "to get the same sparse checkout."
