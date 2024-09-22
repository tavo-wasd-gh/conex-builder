FROM golang:1.22.7
WORKDIR /conex-builder
COPY . .
RUN make
EXPOSE 8080
CMD ["./builder"]
