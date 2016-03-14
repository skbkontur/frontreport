VERSION := $(shell git describe --always --tags --abbrev=0 | tail -c +2)
RELEASE := $(shell git describe --always --tags | awk -F- '{ if ($$2) dot="."} END { printf "1%s%s%s%s\n",dot,$$2,dot,$$3}')
VENDOR := "SKB Kontur"
URL := "https://github.com/skbkontur/csp_reporter"
LICENSE := ""

default: build

build:
	go build -ldflags "-X main.version=$(VERSION)-$(RELEASE)" -o build/csp_reporter

prepare:
	go get github.com/sparrc/gdm
	gdm restore

clean:
	rm -rf build

rpm: clean build
	mkdir -p build/root/usr/local/bin
	mv build/csp_reporter build/root/usr/local/bin/
	fpm -t rpm \
		-s "dir" \
		--description "CSP Reporter" \
		-C build/root \
		--vendor $(VENDOR) \
		--url $(URL) \
		--license $(LICENSE) \
		--name "csp_reporter" \
		--version "$(VERSION)" \
		--iteration "$(RELEASE)" \
		-p build
