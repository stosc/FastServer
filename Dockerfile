

#FROM scratch
FROM golang:latest

WORKDIR $GOPATH/src/fastserver
ADD . $GOPATH/src/fastserver
RUN go get
RUN go build .

ENV KEY="" 
ENV PATH="" 
ENTRYPOINT ["./fastserver -key=$KEY -path=$PATH"]
LABEL Name=fastserver Version=0.0.1
EXPOSE 8899
