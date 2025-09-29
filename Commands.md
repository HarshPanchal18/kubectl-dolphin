# Build commands

```bash

GOOS=linux   GOARCH=amd64 go build -o kubectl-dolphin-linux-amd64
GOOS=darwin  GOARCH=amd64 go build -o kubectl-dolphin-darwin-amd64
GOOS=darwin  GOARCH=arm64 go build -o kubectl-dolphin-darwin-arm64
GOOS=windows GOARCH=amd64 go build -o kubectl-dolphin-windows-amd64.exe

tar -czf kubectl-dolphin-linux-amd64.tar.gz kubectl-dolphin-linux-amd64 LICENSE
tar -czf kubectl-dolphin-darwin-amd64.tar.gz kubectl-dolphin-darwin-amd64 LICENSE
tar -czf kubectl-dolphin-darwin-arm64.tar.gz kubectl-dolphin-darwin-arm64 LICENSE
zip kubectl-dolphin-windows-amd64.zip kubectl-dolphin-windows-amd64.exe LICENSE

sha256sum kubectl-dolphin-*.tar.gz
sha256sum kubectl-dolphin-*.zip
```
