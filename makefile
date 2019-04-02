DEPS= hashing/hashing.go persistentStore/persistentStore.go env/env.go input/input.go logging/logging.go thirdParty
all: server client
server: cmd/server/main.go $(DEPS)
	go build -o $@ $<

client: cmd/client/main.go $(DEPS)
	go build -o $@ $<

thirdParty:
	go get ./...

clean:
	$(RM) client server
