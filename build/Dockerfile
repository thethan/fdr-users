FROM golang:1.14.2-alpine3.11 as base

RUN apk --update add git less openssh && \
    rm -rf /var/lib/apt/lists/* && \
    rm /var/cache/apk/*

WORKDIR $GOPATH/src/github.com/thethan/fdr-users
RUN echo $GOPATH
COPY . $WORKDIR

FROM base as builder
ENV GO111MODULE on
ENV GOPRIVATE "github.com/thethan"
ENV GITHUB_TOKEN "ba062a1cfe39f478cfe35dc44663d084c76dada4"


RUN git config --global url."https://thethan:${GITHUB_TOKEN}@github.com".insteadOf "https://github.com"
RUN mkdir /certs/
RUN go mod vendor
RUN go build -o exec cmd/fdr-users/main.go
RUN go mod download

RUN go build -o fdr-users cmd/fdr-users/main.go


FROM base
EXPOSE 8080 8081 8082
RUN mkdir /certs/
ADD serviceAccountKey.json /certs/serviceAccountKey.json
COPY --from=builder  /go/src/github.com/thethan/fdr-users/exec .
CMD ./exec
COPY --from=builder  /go/src/github.com/thethan/fdr-users/fdr-users .
CMD ./fdr-users
