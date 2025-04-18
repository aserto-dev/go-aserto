SHELL              := $(shell which bash)

NO_COLOR           := \033[0m
OK_COLOR           := \033[32;01m
ERR_COLOR          := \033[31;01m
WARN_COLOR         := \033[36;01m
ATTN_COLOR         := \033[33;01m

GOOS               := $(shell go env GOOS)
GOARCH             := $(shell go env GOARCH)
DOCKER_BUILDKIT    := 1

EXT_DIR            := ${PWD}/.ext
EXT_BIN_DIR        := ${EXT_DIR}/bin
EXT_TMP_DIR        := ${EXT_DIR}/tmp

SVU_VER            := 3.2.3
GOTESTSUM_VER      := 1.12.1
GOLANGCI-LINT_VER  := 2.0.2
GORELEASER_VER     := 2.8.2

RELEASE_TAG		:= $$(${EXT_BIN_DIR}/svu current)

.DEFAULT_GOAL      := build

.PHONY: deps
deps: info install-svu install-golangci-lint install-gotestsum install-goreleaser
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"

.PHONY: build
build:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/goreleaser build --clean --snapshot --single-target

PHONY: go-mod-tidy
go-mod-tidy:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@go work edit -json | jq -r '.Use[].DiskPath' | xargs -I{} bash -c 'cd {} && go mod tidy -v && cd -'

lint:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@${EXT_BIN_DIR}/golangci-lint config path
	@${EXT_BIN_DIR}/golangci-lint config verify
	@go work edit -json | jq -r '.Use[].DiskPath'  | xargs -I{} ${EXT_BIN_DIR}/golangci-lint run {}/... -c .golangci.yaml

test:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@go work edit -json | jq -r '.Use[].DiskPath'  | xargs -I{} ${EXT_BIN_DIR}/gotestsum --format short-verbose -- -count=1 -v {}/...

.PHONY: vault-login
vault-login:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@vault login -method=github token=$$(gh auth token)

.PHONY: info
info:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@echo "GOOS:        ${GOOS}"
	@echo "GOARCH:      ${GOARCH}"
	@echo "EXT_DIR:     ${EXT_DIR}"
	@echo "EXT_BIN_DIR: ${EXT_BIN_DIR}"
	@echo "EXT_TMP_DIR: ${EXT_TMP_DIR}"
	@echo "RELEASE_TAG: ${RELEASE_TAG}"

.PHONY: install-svu
install-svu: ${EXT_BIN_DIR} ${EXT_TMP_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@gh release download v${SVU_VER} --repo https://github.com/caarlos0/svu --pattern "*${GOOS}_all.tar.gz" --output "${EXT_TMP_DIR}/svu.tar.gz" --clobber
	@tar -xvf ${EXT_TMP_DIR}/svu.tar.gz --directory ${EXT_BIN_DIR} svu &> /dev/null
	@chmod +x ${EXT_BIN_DIR}/svu
	@${EXT_BIN_DIR}/svu --version

.PHONY: install-gotestsum
install-gotestsum: ${EXT_TMP_DIR} ${EXT_BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@gh release download v${GOTESTSUM_VER} --repo https://github.com/gotestyourself/gotestsum --pattern "gotestsum_${GOTESTSUM_VER}_${GOOS}_${GOARCH}.tar.gz" --output "${EXT_TMP_DIR}/gotestsum.tar.gz" --clobber
	@tar -xvf ${EXT_TMP_DIR}/gotestsum.tar.gz --directory ${EXT_BIN_DIR} gotestsum &> /dev/null
	@chmod +x ${EXT_BIN_DIR}/gotestsum
	@${EXT_BIN_DIR}/gotestsum --version

.PHONY: install-golangci-lint
install-golangci-lint: ${EXT_TMP_DIR} ${EXT_BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@gh release download v${GOLANGCI-LINT_VER} --repo https://github.com/golangci/golangci-lint --pattern "golangci-lint-${GOLANGCI-LINT_VER}-${GOOS}-${GOARCH}.tar.gz" --output "${EXT_TMP_DIR}/golangci-lint.tar.gz" --clobber
	@tar --strip=1 -xvf ${EXT_TMP_DIR}/golangci-lint.tar.gz --strip-components=1 --directory ${EXT_TMP_DIR} &> /dev/null
	@mv ${EXT_TMP_DIR}/golangci-lint ${EXT_BIN_DIR}/golangci-lint
	@chmod +x ${EXT_BIN_DIR}/golangci-lint
	@${EXT_BIN_DIR}/golangci-lint --version

.PHONY: install-goreleaser
install-goreleaser: ${EXT_TMP_DIR} ${EXT_BIN_DIR}
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@gh release download v${GORELEASER_VER} --repo https://github.com/goreleaser/goreleaser --pattern "goreleaser_$$(uname -s)_$$(uname -m).tar.gz" --output "${EXT_TMP_DIR}/goreleaser.tar.gz" --clobber
	@tar -xvf ${EXT_TMP_DIR}/goreleaser.tar.gz --directory ${EXT_BIN_DIR} goreleaser &> /dev/null
	@chmod +x ${EXT_BIN_DIR}/goreleaser
	@${EXT_BIN_DIR}/goreleaser --version

.PHONY: clean-gen
clean-gen:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@rm -rf ./aserto

.PHONY: clean
clean:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@rm -rf ${EXT_DIR}
	@rm -rf ./dist

${EXT_BIN_DIR}:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@mkdir -p ${EXT_BIN_DIR}

${EXT_TMP_DIR}:
	@echo -e "$(ATTN_COLOR)==> $@ $(NO_COLOR)"
	@mkdir -p ${EXT_TMP_DIR}
