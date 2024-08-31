BIN = builder
SRCDIR = server
SRC = ${SRCDIR}/main.go
GOFILES = ${SRCDIR}/go.sum ${SRCDIR}/go.mod
GOMODS = github.com/joho/godotenv github.com/lib/pq

all: ${BIN}

${BIN}: ${SRC} ${GOFILES}
	(cd ${SRCDIR} && go build -o ../${BIN})

${GOFILES}:
	(cd ${SRCDIR} && go mod init ${BIN})
	(cd ${SRCDIR} && go get ${GOMODS})

start: ${BIN}
	@./$< &

stop:
	-@pkill -SIGTERM ${BIN} || true

restart: stop start

clean-all: clean clean-mods

clean:
	rm -f ${BIN}

clean-mods:
	rm -f ${SRCDIR}/go.*
