BIN = builder
SRC = main.go
GOFILES = go.sum go.mod
GOMODS = github.com/joho/godotenv

all: ${BIN}

${BIN}: ${SRC} ${GOFILES}
	go build -o $@

${GOFILES}:
	go mod init ${BIN}
	go get ${GOMODS}

start: ${BIN}
	@./$< &

stop:
	-@pkill -SIGTERM ${BIN} || true

restart: stop start

clean-all: clean clean-mods

clean:
	rm -f ${BIN}

clean-mods:
	rm -f go.*
