VERSION := $(shell git describe --always --tags --abbrev=0 | tail -c +2)
RELEASE := $(shell git describe --always --tags | awk -F- '{ if ($$2) dot="."} END { printf "1%s%s%s%s\n",dot,$$2,dot,$$3}')
VENDOR := "SKB Kontur"
URL := "https://github.com/skbkontur/frontreport"
LICENSE := "BSD"

default: clean prepare test build rpm

build:
	mkdir build
	cd cmd/frontreport && go build -ldflags "-X main.version=$(VERSION)-$(RELEASE)" -o ../../build/frontreport

prepare:
	go get github.com/kardianos/govendor
	govendor sync

test: prepare
	echo "No tests"

clean:
	rm -rf build

rpm: clean build
	mkdir -p build/root/usr/bin
	cp build/frontreport build/root/usr/bin/
	fpm -t rpm \
		-s dir \
		--description "CSP/HPKP Report Collector" \
		-C build/root \
		--vendor $(VENDOR) \
		--url $(URL) \
		--license $(LICENSE) \
		--name frontreport \
		--version "$(VERSION)" \
		--iteration "$(RELEASE)" \
		-p build
