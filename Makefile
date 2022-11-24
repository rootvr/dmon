DMON_GO := go

DMON_MAIN := supervisor.go
DMON_OBJ := dmon

PREFIX := .
DESTDIR := bin

# sockshop
DMON_WIRESHARK_IFACES := weave
DMON_DOCKER_IFACES := weave
# robot-shop
# DMON_WIRESHARK_IFACES := robot-shop
# DMON_DOCKER_IFACES := robotshop_robot-shop

DMON_DOCKER_TIMER := 5
DMON_REDIS_NETLOC := localhost:6379

DMON_ARGS := -i $(DMON_WIRESHARK_IFACES) -n $(DMON_DOCKER_IFACES) -t $(DMON_DOCKER_TIMER) -r $(DMON_REDIS_NETLOC)

DMON_DOCKER := docker
DMON_IMAGE := rootvr/dmon

.SILENT: help
.PHONY: help # print help
help:
	grep '^.PHONY: .* #' $(firstword $(MAKEFILE_LIST)) |\
	sed 's/\.PHONY: \(.*\) # \(.*\)/\1 # \2/' |\
	awk 'BEGIN {FS = "#"}; {printf "%-20s %s\n", $$1, $$2}' 

.PHONY: build # compile dmon
build:
	$(DMON_GO) build -o $(PREFIX)/$(DESTDIR)/$(DMON_OBJ)

.PHONY: clean # clean and remove dmon binary
clean:
	$(DMON_GO) clean
	rm -f $(PREFIX)/$(DESTDIR)/$(DMON_OBJ)

.PHONY: run # compile and run dmon
run: build
	clear
	$(DMON_GO) run $(DMON_MAIN) $(DMON_ARGS)

.PHONY: install # install dmon into ~/.local/bin
install:
	install -D $(PREFIX)/$(DESTDIR)/$(DMON_OBJ) ${HOME}/.local/bin/$(DMON_OBJ)

.PHONY: uninstall # unininstall dmon from ~/.local/bin
uninstall:
	rm -f ${HOME}/.local/bin/$(DMON_OBJ)

.PHONY: deps # resolve dependencies
deps:
	$(DMON_GO) mod tidy

.SILENT: fmt
.PHONY: fmt # format all
fmt:
	$(DMON_GO) fmt ./...

.PHONY: gen # generate dmon docker image
gen:
	$(DMON_DOCKER) build -f docker/Dockerfile -t $(DMON_IMAGE) .

.PHONY: rmi # delete dmon docker image
rmi:
	$(DMON_DOCKER) rmi -f $(DMON_IMAGE)
