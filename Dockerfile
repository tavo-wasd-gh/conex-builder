FROM golang:1.19
WORKDIR /conex-builder
COPY . .
RUN make
EXPOSE 8080
CMD ["./builder"]
