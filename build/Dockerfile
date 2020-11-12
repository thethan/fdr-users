FROM golang:1.15.2-alpine3.12 as base

RUN apk --update add git less openssh && \
    rm -rf /var/lib/apt/lists/* && \
    rm /var/cache/apk/*

WORKDIR $GOPATH/src/github.com/thethan/fdr-users
RUN echo $GOPATH
COPY . $WORKDIR

FROM base as builder
ENV GO111MODULE on
ENV GOPRIVATE "github.com/thethan"
ENV GITHUB_TOKEN "a7de91c4fa8d2c2e53c338d60a1850b3037f5c3e"


RUN git config --global url."https://thethan:${GITHUB_TOKEN}@github.com".insteadOf "https://github.com"
RUN mkdir /certs/
RUN go mod download

RUN go build -o fdr-users cmd/fdr-users/main.go


FROM base
EXPOSE 8080 8081 8082
RUN mkdir /certs/
ADD serviceAccountKey.json /certs/serviceAccountKey.json
COPY --from=builder  /go/src/github.com/thethan/fdr-users/fdr-users .
CMD ./fdr-users
