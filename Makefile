
OS := linux
ARCH := arm
ARM := 6

BEDROOM := bedroom
UTILITYROOM := 192.168.1.17

.PHONY: bedroom utilityroom clean deploy-bedroom deploy-utilityroom

all: bedroom utilityroom

bedroom:
	GOOS=$(OS) GOARCH=$(ARCH) GOARM=$(ARM) go build -o bedroom ./cmd/bedroom/

utilityroom:
	GOOS=$(OS) GOARCH=$(ARCH) GOARM=$(ARM) go build -o utilityroom ./cmd/utilityroom/


clean:
	rm ./bedroom ./gpiotest ./utilityroom

deploy-bedroom: bedroom
	/usr/bin/scp -i ~/.ssh/id_ed25519 ./bedroom kit@$(BEDROOM):/home/kit/bin/bedroom

deploy-utilityroom: utilityroom
	/usr/bin/scp -i ~/.ssh/id_ed25519 ./utilityroom kit@$(UTILITYROOM):/home/kit/bin/utilityroom




