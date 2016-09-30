VERSION := $(shell git describe --always --tags --abbrev=0 | tail -c +2)
RELEASE := $(shell git describe --always --tags | awk -F- '{ if ($$2) dot="."} END { printf "1%s%s%s%s\n",dot,$$2,dot,$$3}')
VENDOR := "SKB Kontur"
URL := "https://github.com/skbkontur/cspreport"
LICENSE := "BSD"

default: clean prepare test build rpm

build:
	mkdir build
	cd cmd/cspreport && go build -ldflags "-X main.version=$(VERSION)-$(RELEASE)" -o ../../build/cspreport

prepare:
	go get github.com/kardianos/govendor
	govendor sync

test: prepare
	echo "No tests"

clean:
	rm -rf build

rpm: clean build
	mkdir -p build/root/usr/bin
	cp build/cspreport build/root/usr/bin/
	fpm -t rpm \
		-s dir \
		--description "CSP/HPKP Report Collector" \
		-C build/root \
		--vendor $(VENDOR) \
		--url $(URL) \
		--license $(LICENSE) \
		--name cspreport \
		--version "$(VERSION)" \
		--iteration "$(RELEASE)" \
		-p build
