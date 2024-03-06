#!/bin/zsh

# Finalize O/S
sudo apt update -y
sudo apt upgrade -y
sudo apt install -y vim inotify-tools iputils-ping socat
sudo chown -R vscode:vscode /workspaces/

# Install Go tools
go install mvdan.cc/gofumpt@latest

# Setup zsh environment
alias ll='ls -alF'
alias pj='npx projen'

echo "alias ll='ls -laFh'" >> /home/vscode/.zshrc
echo "alias pj='npx projen'" >> /home/vscode/.zshrc
echo "alias gotest='while inotifywait -r -e close_write ./ ; do go test ./... -v -cover; done'" >> /home/vscode/.zshrc
echo "alias xclip='socat - tcp:host.docker.internal:8121'" >> /home/vscode/.zshrc
echo "alias printgo='$(git rev-parse --show-toplevel)/.scripts/print_go_files.sh'" >> /home/vscode/.zshrc


