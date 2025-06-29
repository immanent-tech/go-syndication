#!/usr/bin/env bash

set -e

# Install additional packages.
sudo apt-get update && export DEBIAN_FRONTEND=noninteractive &&
    sudo apt-get -y install --no-install-recommends micro pre-commit &&
    sudo apt-get -y autoremove && sudo apt-get -y clean && sudo rm -rf /var/lib/apt/lists/*

# Add starship to fish shell.
mkdir -p ~/.config/fish/conf.d
echo "starship init fish | source" > ~/.config/fish/conf.d/999-starship.fish
# Add starship to bash shell.
echo 'eval "$(starship init bash)"' >>~/.bashrc

exit 0
