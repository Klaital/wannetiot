
pi:
	GOOS=linux GOARCH=arm GOARM=6 go build ./cmd/bedroom/

all:
	go build ./cmd/bedroom/

clean:
	rm ./bedroom

deploy: pi
	/usr/bin/scp -i ~/.ssh/id_builder_ed25519 ./bedroom pi@192.168.1.6:/home/pi/bin/
