lint_cmd: 
	golangci-lint run -v || true
lint: lint_cmd
