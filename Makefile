# DrawExact Observe build and verification tooling.
#
# Developer entry points:
#   make test          Quick, frontend-independent Go tests.
#   make verify        Full frontend build plus whole-module Go checks.
#
# Frontend build pipeline:
#   install-ui-deps -> build-ui -> stage-ui-assets -> verify
#
# Frontend commands use the Node.js major version specified in ui/.nvmrc.
# Generated frontend assets are embedded into the Go application at build time.


# Everyday backend tests for internal packages only. Does not need Node.js,
# Vite, or a staged static/ frontend tree.
.PHONY: test
test:
	go test -count=1 ./internal/...

# Full build and test verification.
# Installs UI dependencies, builds the Svelte app, and stages the contents
# of ui/dist/ into static/.
# Runs uncached Go tests across the entire module and checks that the root
# package builds successfully with its embedded static/ assets.
# The compiled executable is discarded.
.PHONY: verify
verify: stage-ui-assets
	go test -count=1 ./...
	go build -o /dev/null .

# Helper macro for executing frontend commands with the Node.js major version
# specified in ui/.nvmrc.
# Uses the active Node installation if its major version matches; otherwise
# attempts to select the required version through nvm.
# Fails if the required version is unavailable; does not install Node.js.
# Any version switching is local to the recipe's shell.
define with-required-node
	expected=$$(tr -d '[:space:]' < ui/.nvmrc); \
	if command -v node >/dev/null 2>&1; then \
		major=$$(node -p "process.versions.node.split('.')[0]"); \
		if [ "$$major" = "$$expected" ]; then \
			$(1); exit $$?; \
		fi; \
	fi; \
	if [ -s "$${NVM_DIR:-$$HOME/.nvm}/nvm.sh" ]; then \
		. "$${NVM_DIR:-$$HOME/.nvm}/nvm.sh" && nvm use "$$expected" && $(1); \
	else \
		echo "need Node $$expected (see ui/.nvmrc)"; exit 1; \
	fi
endef

# Install frontend dependencies from the lockfile when present.
# Replaces any existing node_modules/ directory with a clean installation.
# Uses the Node.js version specified in ui/.nvmrc.
.PHONY: install-ui-deps
install-ui-deps:
	@$(call with-required-node,cd ui && npm ci)

# Produce a production build of the Svelte/Vite frontend in ui/dist/.
# Performs a clean installation of frontend dependencies before building.
# The resulting assets are ready for staging into the Go application.
.PHONY: build-ui
build-ui: install-ui-deps
	@$(call with-required-node,cd ui && npm run build)

# Prepare the production frontend assets for embedding in the Go application
# via //go:embed.
# Builds the frontend, then completely replaces static/ with the contents
# of ui/dist/, ensuring no obsolete assets remain.
# static/ contains generated assets and is gitignored.
.PHONY: stage-ui-assets
stage-ui-assets: build-ui
	rm -rf static
	mkdir -p static
	cp -a ui/dist/. static/

# Installs golines and rewrites Go sources to an 80-character line length.
# Run locally when you want formatting; not invoked by test or verify.
.PHONY: golines
golines:
	go install github.com/segmentio/golines@latest
	golines -w -m80 .

