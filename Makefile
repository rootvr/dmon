DMON_GO := go

DMON_MAIN := ./supervisor.go
DMON_OBJ := dmon

PREFIX := .
DESTDIR := bin

IPREFIX := ${HOME}
IDESTDIR := .local/bin

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
DMON_DOCKERFILE := ./docker/Dockerfile
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
run:
	clear
	$(DMON_GO) run $(DMON_MAIN) $(DMON_ARGS)

.PHONY: install # install dmon into ~/.local/bin
install:
	install -D $(PREFIX)/$(DESTDIR)/$(DMON_OBJ) $(IPREFIX)/$(IDESTDIR)/$(DMON_OBJ)

.PHONY: uninstall # unininstall dmon from ~/.local/bin
uninstall:
	rm -f $(IPREFIX)/$(IDESTDIR)/$(DMON_OBJ)

.PHONY: deps # resolve dependencies
deps:
	$(DMON_GO) mod tidy

.SILENT: format
.PHONY: format # format all
format:
	$(DMON_GO) fmt ./...

.PHONY: genimage # generate dmon docker image
genimage:
	$(DMON_DOCKER) build -f $(DMON_DOCKERFILE) -t $(DMON_IMAGE) .

.PHONY: delimage # delete dmon docker image
delimage:
	$(DMON_DOCKER) rmi -f $(DMON_IMAGE)
