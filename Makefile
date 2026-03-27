.PHONY: test ctest covdir coverage docs linter qtest clean dep release license
APP_VERSION:=$(shell cat VERSION | head -1)
GIT_COMMIT:=$(shell git describe --dirty --always)
GIT_BRANCH:=$(shell git rev-parse --abbrev-ref HEAD -- | head -1)
BUILD_USER:=$(shell whoami)
BUILD_DATE:=$(shell date +"%Y-%m-%d")
BINARY:="versioned"
VERBOSE:=-v
ifdef TEST
	TEST:="-run ${TEST}"
endif

.PHONY: all
all: info build
	@echo "DEBUG: completed $@"

.PHONY: info
info:
	@echo "Version: $(APP_VERSION), Branch: $(GIT_BRANCH), Revision: $(GIT_COMMIT)"
	@echo "Build on $(BUILD_DATE) by $(BUILD_USER)"

.PHONY: build
build:
	@echo "DEBUG: started $@"
	@mkdir -p bin/
	@CGO_ENABLED=0 go build -o bin/$(BINARY) $(VERBOSE) \
		-ldflags="-w -s \
		-X main.gitBranch=$(GIT_BRANCH) \
		-X main.gitCommit=$(GIT_COMMIT) \
		-X main.buildUser=$(BUILD_USER) \
		-X main.buildDate=$(BUILD_DATE)" \
		-gcflags="all=-trimpath=$(GOPATH)/src" \
		-asmflags="all=-trimpath $(GOPATH)/src" cmd/$(BINARY)/*.go
	@chmod +x bin/$(BINARY)
	@./bin/$(BINARY) --version
	@echo "DEBUG: completed $@"

.PHONY: linter
linter:
	@echo "Running lint checks"
	@golint *.go
	@golint cmd/$(BINARY)/*.go
	@echo "DEBUG: completed $@"

.PHONY: test
test: covdir linter
	@go test $(VERBOSE) -coverprofile=.coverage/coverage.out ./*.go
	@echo "DEBUG: completed $@"

.PHONY: ctest
ctest: covdir linter
	@richgo version || go install github.com/kyoh86/richgo@latest
	@time richgo test $(VERBOSE) $(TEST) -coverprofile=.coverage/coverage.out ./*.go

.PHONY: covdir
covdir:
	@echo "Creating .coverage/ directory"
	@mkdir -p .coverage

.PHONY: coverage
coverage:
	@#go tool cover -help
	@go tool cover -html=.coverage/coverage.out -o .coverage/coverage.html
	@go test -covermode=count -coverprofile=.coverage/coverage.out ./*.go
	@go tool cover -func=.coverage/coverage.out | grep -v "100.0"
	@echo "DEBUG: completed $@"

.PHONY: license
license:
	@for f in `find ./ -type f -name '*.go'`; do ./bin/versioned -addlicense -copyright="Paul Greenberg (greenpau@outlook.com)" -year=2020 -filepath=$$f; done

.PHONY: docs
docs:
	@mkdir -p .doc
	@go doc -all > .doc/index.txt
	@cat .doc/index.txt

.PHONY: clean
clean:
	@rm -rf .doc
	@rm -rf .coverage
	@rm -rf bin/

.PHONY: qtest
qtest:
	@echo "Perform quick tests ..."
	@#go test -v -run TestVersioned *.go

.PHONY: dep
dep:
	@echo "Making dependencies check ..."
	@golint || go install golang.org/x/lint/golint@latest
	@echo "DEBUG: completed $@"

.PHONY: release-git-check
release-git-check:
	@echo "DEBUG: started $@"
	@go mod tidy;
	@go mod verify;
	@if [ $(GIT_BRANCH) != "main" ]; then echo "cannot release to non-main branch $(GIT_BRANCH)" && false; fi
	@git diff-index --quiet HEAD -- || ( echo "git directory is dirty, commit changes first" && false )
	@echo "DEBUG: completed $@"

.PHONY: release-update-version
release-update-version:
	@echo "DEBUG: started $@"
	@./bin/$(BINARY) -patch
	@./bin/$(BINARY) -prerelease -sync cmd/$(BINARY)/main.go
	@git add VERSION cmd/$(BINARY)/main.go

.PHONY: release-git-commit
release-git-commit:
	@echo "DEBUG: started $@"
	@git commit -m "released v`cat VERSION | head -1`"
	@git tag -a v`cat VERSION | head -1` -m "v`cat VERSION | head -1`"
	@git push
	@git push --tags
	@echo "If necessary, run the following commands:"
	@echo "  git push --delete origin v$(APP_VERSION)"
	@echo "  git tag --delete v$(APP_VERSION)"
	@echo "DEBUG: completed $@"

.PHONY: release
release: release-git-check build release-update-version release-git-commit
	@echo "DEBUG: completed $@"
