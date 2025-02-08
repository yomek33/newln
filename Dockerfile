FROM golang:1.23

WORKDIR /app
COPY . .

# 認証情報ファイルをコンテナにコピー
COPY newln-448314-ed4973430f26.json /app/newln-448314-ed4973430f26.json

RUN go mod download
RUN apt-get update && apt-get install -y git
RUN git clone https://github.com/cosmtrek/air.git /tmp/air && \
    cd /tmp/air && \
    go build -o /go/bin/air
RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list \
    && curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg \
    && apt-get update -y \
    && apt-get install google-cloud-sdk -y

ENV PATH="/go/bin:${PATH}"
# 認証情報のパス
ENV GOOGLE_APPLICATION_CREDENTIALS="/app/newln-448314-ed4973430f26.json"
ENV CLOUDSDK_CORE_PROJECT="newln-448314"
CMD ["air", "-c", ".air.toml"]
