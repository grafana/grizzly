FROM golang:latest
WORKDIR /app
COPY . .
RUN make install
ENTRYPOINT ["./grr"]
