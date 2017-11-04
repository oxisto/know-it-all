FROM golang:1.9

WORKDIR /go/src/github.com/oxisto/know-it-all

# install dep utility
RUN go get -u github.com/golang/dep/cmd/dep

# copy dependency information and fetch them
COPY Gopkg.* ./
RUN dep ensure --vendor-only

# copy sources
COPY . .

# build and install
RUN go-wrapper install

CMD ["go-wrapper", "run"]
