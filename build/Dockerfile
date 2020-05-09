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
ENV GITHUB_TOKEN "1c909977e86be4c8b2a9ab4a495731787f2713c2"

RUN git config --global url."https://thethan:${GITHUB_TOKEN}@github.com".insteadOf "https://github.com"
RUN mkdir /certs/
ADD serviceAccountKey.json /certs/serviceAccountKey.json
RUN go mod vendor
RUN go build -o exec cmd/fdr-users/main.go


FROM base
EXPOSE 80 8080 8082
COPY --from=builder  /go/src/github.com/thethan/fdr-users/exec .
CMD ./exec

