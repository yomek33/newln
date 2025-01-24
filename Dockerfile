FROM golang:1.23

WORKDIR /app
COPY . .
RUN go mod download
RUN apt-get update && apt-get install -y git
RUN git clone https://github.com/cosmtrek/air.git /tmp/air && \
    cd /tmp/air && \
    go build -o /go/bin/air

ENV PATH="/go/bin:${PATH}"
CMD ["air", "-c", ".air.toml"]
