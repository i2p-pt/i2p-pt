
all: server client

server:
	go build -o i2p-server-bin ./i2p-server

client:
	go build -o i2p-client-bin ./i2p-client