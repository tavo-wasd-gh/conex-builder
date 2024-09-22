BIN = builder
SRCDIR = server
SRC = ${SRCDIR}/main.go \
      ${SRCDIR}/paypal.go \
      ${SRCDIR}/db.go \
      ${SRCDIR}/auth.go \
      ${SRCDIR}/bucket.go \

GO = go
GOFILES = ${SRCDIR}/go.sum ${SRCDIR}/go.mod
GOMODS = github.com/joho/godotenv \
	 github.com/lib/pq \
	 gopkg.in/gomail.v2 \
	 github.com/aws/aws-sdk-go-v2/aws \
	 github.com/aws/aws-sdk-go-v2/config \
	 github.com/aws/aws-sdk-go-v2/credentials \
	 github.com/aws/aws-sdk-go-v2/service/s3 \

all: ${BIN} fmt

${BIN}: ${SRC} ${GOFILES}
	(cd ${SRCDIR} && ${GO} build -o ../${BIN})

fmt: ${SRC}
	@diff=$$(gofmt -d $^); \
	if [ -n "$$diff" ]; then \
		printf '%s\n' "$$diff"; \
		exit 1; \
	fi

${GOFILES}:
	(cd ${SRCDIR} && ${GO} mod init ${BIN})
	(cd ${SRCDIR} && ${GO} get ${GOMODS})

start: ${BIN}
	@./$< &

stop:
	-@pkill -SIGTERM ${BIN} || true

restart: stop start

clean-all: clean clean-mods

clean:
	rm -f ${BIN}

clean-mods:
	go clean -modcache
	rm -f ${SRCDIR}/go.*
