
OS := linux
ARCH := arm
ARM := 6

.PHONY: bedroom utilityroom clean deploy-bedroom deploy-utilityroom

all: bedroom utilityroom

bedroom:
	GOOS=$(OS) GOARCH=$(ARCH) GOARM=$(ARM) go build -o bedroom ./cmd/bedroom/

utilityroom:
	GOOS=$(OS) GOARCH=$(ARCH) GOARM=$(ARM) go build -o utilityroom ./cmd/utilityroom/


clean:
	rm ./bedroom ./gpiotest ./utilityroom

deploy-bedroom: bedroom
	/usr/bin/scp -i ~/.ssh/id_builder_ed25519 ./bedroom pi@bedroom:/home/pi/bin/bedroom

deploy-utilityroom: utilityroom
	/usr/bin/scp -i ~/.ssh/id_builder_ed25519 ./utilityroom pi@192.168.1.11:/home/pi/bin/utilityroom




