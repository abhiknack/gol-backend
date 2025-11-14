#!/bin/bash
# Fix Go module cache permissions in dev container

echo "Fixing Go module cache permissions..."
docker exec -u root devcontainer-app-1 chown -R vscode:vscode /go
echo "✓ Permissions fixed!"
echo ""
echo "You can now run: go run cmd/server/main.go"
