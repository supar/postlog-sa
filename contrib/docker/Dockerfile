FROM debian:wheezy
MAINTAINER  Pavel "paulrez@gmail.com"

# Build spam-bug package
#

# Get rid of the debconf messages
ENV DEBIAN_FRONTEND noninteractive

# Add backport
RUN echo "deb http://http.debian.net/debian wheezy-backports main" >> /etc/apt/sources.list

# Update packages and install the ones that are needed
RUN apt-get update -y
RUN apt-get upgrade -y
RUN apt-get install -y curl ca-certificates gcc libc6-dev \
                       bzr git-core mercurial locales make \
                       build-essential debhelper devscripts \
                       fakeroot \
                       --no-install-recommends \
                       && rm -rf /var/lib/apt/lists/*

RUN echo "Europe/Moscow" > /etc/timezone && dpkg-reconfigure --frontend noninteractive tzdata
ENV LANG C.UTF-8

# Go install section
RUN curl -sSL https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz \
    | tar -v -C /usr/local -xz

ENV PATH /usr/local/go/bin:$PATH
RUN mkdir -p /go/src
ENV GOPATH /go
ENV PATH /go/bin:$PATH

# Install beegoo dependency
RUN go get github.com/astaxie/beego && \
    go get gopkg.in/ini.v1 && \
    go get github.com/hpcloud/tail && \
    go get github.com/smartystreets/goconvey/convey && \
    go get github.com/go-sql-driver/mysql && \
    go get gopkg.in/DATA-DOG/go-sqlmock.v1

ADD run.sh /run.sh
ENTRYPOINT [ "/run.sh" ]
