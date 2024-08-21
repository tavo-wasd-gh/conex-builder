BIN = builder
SRC = main.go
GOFILES = go.sum go.mod
GOMODS = github.com/joho/godotenv github.com/lib/pq

all: ${BIN}

${BIN}: ${SRC} ${GOFILES}
	go build -o builder

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
