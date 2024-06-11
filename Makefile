# run this command once, when you work on this project for the first time.
init:
	@echo "== 👩‍🌾 ci init =="
	@if ! command -v node > /dev/null; then \
		@echo "== Node.js is not installed, installing Node.js... =="; \
		brew install node; \
	fi;
	@if ! command -v pre-commit > /dev/null; then \
		@echo "== pre-commit is not installed, installing pre-commit... =="; \
		brew install pre-commit; \
	fi;
	@if ! command -v golangci-lint > /dev/null; then \
		@echo "== golangci-lint is not installed, installing golangci-lint... =="; \
		brew install golangci-lint; \
	fi;

	brew upgrade golangci-lint

	@echo "== pre-commit setup =="
	pre-commit install

	@echo "== install hook =="
	$(MAKE) precommit.rehooks

# installs the pre-commit hooks defined in the .pre-commit-config.yaml, specifically installs the commit-msg hook. \
The commit-msg hook is a special type of pre-commit hook that runs after you write your commit message \
but before the commit is finalized.
precommit.rehooks:
	pre-commit autoupdate
	pre-commit install --install-hooks
	pre-commit install --hook-type commit-msg


lint.fix:
	@golangci-lint run --fix
