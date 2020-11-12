FROM golang:1.14.2-alpine3.11 as base

RUN apk --update add git less openssh && \
    rm -rf /var/lib/apt/lists/* && \
    rm /var/cache/apk/*

WORKDIR $GOPATH/src/github.com/thethan/fdr-players-import
RUN echo $GOPATH
COPY . $WORKDIR

FROM base as builder
ENV GO111MODULE on
ENV GOPRIVATE "github.com/thethan"
ENV GITHUB_TOKEN "29c653186c77dcffafc70025b73daaf767ee9d6a"


RUN git config --global url."https://thethan:${GITHUB_TOKEN}@github.com".insteadOf "https://github.com"
RUN mkdir /certs/
RUN go mod vendor
RUN go build -o exec cmd/fdr-players-import/main.go
RUN go mod download

RUN go build -o fdr-players cmd/fdr-players-import/main.go


FROM base
EXPOSE 8080 8081 8082
RUN mkdir /certs/
ADD serviceAccountKey.json /certs/serviceAccountKey.json
COPY --from=builder  /go/src/github.com/thethan/fdr-players-import/fdr-players .
CMD ./fdr-users
