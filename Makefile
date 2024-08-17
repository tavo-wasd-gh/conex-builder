BIN = builder
SRC = main.go
GOMODS = github.com/joho/godotenv

all: ${BIN}

${BIN}: ${SRC}
	go build -o $@

run: ${BIN}
	@./$< &

stop:
	-@pkill -SIGTERM ${BIN} || true

restart: stop run

bootstrap:
	go mod init ${BIN}
	go get ${GOMODS}

clean:
	rm -f ${BIN}

clean-mods:
	rm -f go.*
