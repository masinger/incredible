define build
	GOOS=$1 GOARCH=$2 CGO_ENABLED=0 go build -o bin/incredible-$1-$2$3
endef

clean:
	rm -f bin/*

linux-amd64:
	$(call build,linux,amd64)
linux-arm64:
	$(call build,linux,arm64)
linux: linux-amd64 linux-arm64

windows-amd64:
	$(call build,windows,amd64,.exe)
windows: windows-amd64

darwin-arm64:
	$(call build,darwin,arm64)
darwin-amd64:
	$(call build,darwin,amd64)
darwin: darwin-amd64 darwin-arm64

freebsd-amd64:
	$(call build,freebsd,amd64)
freebsd-arm:
	$(call build,freebsd,arm)
freebsd: freebsd-amd64 freebsd-arm

openbsd-amd64:
	$(call build,openbsd,amd64)
openbsd-arm:
	$(call build,openbsd,arm)
openbsd: openbsd-amd64 openbsd-arm

netbsd-amd64:
	$(call buid,netbsd,amd64)
netbsd-arm:
	$(call build,netbsd,arm)
netbsd: netbsd-amd64 netbsd-arm

build: linux windows darwin freebsd openbsd netbsd

