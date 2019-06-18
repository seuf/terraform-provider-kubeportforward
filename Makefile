VERSION = v0.0.1
LDFLAGS = -ldflags '-s -w'
GOARCH = amd64
linux: export GOOS=linux
linux_arm: export GOOS=linux
linux_arm: export GOARCH=arm
linux_arm: export GOARM=6
linux_arm64: export GOOS=linux
linux_arm64: export GOARCH=arm64
darwin: export GOOS=darwin
windows: export GOOS=windows

all: linux linux_arm linux_arm64 darwin windows

linux:
	mkdir -p release
	rm -f terraform-provider-kubeportforward_${VERSION} release/terraform-provider-kubeportforward_${VERSION}_${GOOS}_${GOARCH}.zip
	go build $(LDFLAGS) -o terraform-provider-kubeportforward_${VERSION}
	chmod +x terraform-provider-kubeportforward_${VERSION}
	zip release/terraform-provider-kubeportforward_${VERSION}_${GOOS}_${GOARCH}.zip terraform-provider-kubeportforward_${VERSION}

linux_arm:
	mkdir -p release
	rm -f terraform-provider-kubeportforward_${VERSION} release/terraform-provider-kubeportforward_${VERSION}_${GOOS}_${GOARCH}.zip
	go build $(LDFLAGS) -o terraform-provider-kubeportforward_${VERSION}
	chmod +x terraform-provider-kubeportforward_${VERSION}
	zip release/terraform-provider-kubeportforward_${VERSION}_${GOOS}_${GOARCH}.zip terraform-provider-kubeportforward_${VERSION}

linux_arm64:
	mkdir -p release
	rm -f terraform-provider-kubeportforward_${VERSION} release/terraform-provider-kubeportforward_${VERSION}_${GOOS}_${GOARCH}.zip
	go build $(LDFLAGS) -o terraform-provider-kubeportforward_${VERSION}
	chmod +x terraform-provider-kubeportforward_${VERSION}
	zip release/terraform-provider-kubeportforward_${VERSION}_${GOOS}_${GOARCH}.zip terraform-provider-kubeportforward_${VERSION}

darwin:
	mkdir -p release
	rm -f terraform-provider-kubeportforward_${VERSION} release/terraform-provider-kubeportforward_${VERSION}_${GOOS}_${GOARCH}.zip
	go build $(LDFLAGS) -o terraform-provider-kubeportforward_${VERSION}
	chmod +x terraform-provider-kubeportforward_${VERSION}
	zip release/terraform-provider-kubeportforward_${VERSION}_${GOOS}_${GOARCH}.zip terraform-provider-kubeportforward_${VERSION}

windows:
	mkdir -p release
	rm -f terraform-provider-kubeportforward_${VERSION}.exe release/terraform-provider-kubeportforward_${VERSION}_${GOOS}_${GOARCH}.zip
	go build $(LDFLAGS) -o terraform-provider-kubeportforward_${VERSION}.exe
	zip release/terraform-provider-kubeportforward_${VERSION}_${GOOS}_${GOARCH}.zip terraform-provider-kubeportforward_${VERSION}.exe

.PHONY: clean
clean:
	rm -rf release
	rm -f terraform-provider-kubeportforward terraform-provider-kubeportforward.exe
