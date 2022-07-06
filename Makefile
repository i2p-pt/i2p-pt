
all: server client index

server:
	go build -o i2p-server-bin ./i2p-server

client:
	go build -o i2p-client-bin ./i2p-client

index:
	pandoc -s -t html \
		-c ./css/style.css \
		--highlight-style=tango \
		--metadata title="I2P Pluggable Transport" -o index.html README.md