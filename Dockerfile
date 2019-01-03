FROM golang:1.11.2-alpine3.8
MAINTAINER M. - Karan Bhomia

ENV SOURCES /go/src/github.com/karanbhomiagit/order-service/

RUN apk update -qq && apk add git

COPY . ${SOURCES}
RUN go get googlemaps.github.io/maps
RUN go get gopkg.in/mgo.v2
RUN go get github.com/stretchr/testify
RUN cd ${SOURCES} && CGO_ENABLED=0 go install

ENV PORT 8080
ENV PAGE_SIZE 10
ENV GOOGLE_API_KEY <Your API Key>
ENV MONGODB_URL <Mongo DB URL>
ENV DATABASE_NAME order-service-db

EXPOSE 3000

ENTRYPOINT order-service