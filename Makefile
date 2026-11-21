.PHONY: check css browser security sbom demo icons-free icons-pro demo-pro release-binaries

check: css
	go test ./...

css:
	npm run check:css
	npm run build:demo:css

icons-free:
	node scripts/build-icons.js free

icons-pro:
	node scripts/build-icons.js pro

browser: css
	npm run test:browser

demo: css
	go run ./cmd/vyview

demo-pro: css icons-pro
	go run -tags fapro ./cmd/vyview

release-binaries: css
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/vyview-linux-arm64 ./cmd/vyview
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/vyview-linux-amd64 ./cmd/vyview

security:
	go tool govulncheck ./...
	go tool govulncheck github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod
	npm audit --audit-level=high

sbom:
	mkdir -p dist
	go tool cyclonedx-gomod mod -json -output dist/go-sbom.json
	npx --no-install cyclonedx-npm --output-file dist/npm-sbom.json
