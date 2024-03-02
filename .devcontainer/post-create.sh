#!/bin/zsh

# Finalize O/S
sudo apt update -y
sudo apt upgrade -y
sudo apt install -y vim inotify-tools iputils-ping
sudo chown -R vscode:vscode /workspaces/

# Install Go tools
go install mvdan.cc/gofumpt@latest

# Setup zsh environment
alias ll='ls -alF'
alias pj='npx projen'

echo "alias ll='ls -laFh'" >> /home/vscode/.zshrc
echo "alias pj='npx projen'" >> /home/vscode/.zshrc
echo "alias gotest='while inotifywait -r -e close_write ./ ; do go test ./... -v -cover; done'" >> /home/vscode/.zshrc

