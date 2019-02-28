FROM golang
#ENV GIN_MODE release
#ENV GOPATH /home/ubuntu/wordx

ADD ./ /go/
WORKDIR /go/src/github.com/heroiclabs/nakama
RUN go get -u github.com/golang/dep/...
#CMD ["go/src/github.com/heroiclabs/nakama"]
RUN  dep ensure

#ENV GOBIN /wbserv/bin
#COPY ./wb.crt /wbserv/bin
#COPY ./wb.key /wbserv/bin
# Copy the local package files to the container's workspace..
#COPY ./ /wbserv/src
EXPOSE 9090 7348 7351 7349 7350 7352

RUN go install github.com/heroiclabs/nakama/
#RUN go/bin/nakama --migrate_up

#CMD ["/go/bin/nakama migrate up"]
#RUN /go/bin/nakama migrate up

ENTRYPOINT /go/bin/nakama $OPT

#EXPOSE 9090, 7348, 7351, 7349, 7350, 735
