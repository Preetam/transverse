FROM ubuntu:16.04

# gcc for cgo
RUN apt-get update && apt-get install -y --no-install-recommends \
#		g++ \
#		gcc \
#		libc6-dev \
		make \
		pkg-config \
		curl \
	&& rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.10
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 b5a64335f1490277b585832d1f6c7f8c6c11206cba5cd3f771dcb87b98ad1a33

RUN curl -k -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
	&& echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
	&& tar -C /usr/local -xzf golang.tar.gz \
	&& rm golang.tar.gz

RUN apt-get update && apt-get install -y npm nodejs-legacy python-pip

RUN npm install -g uglify-js clean-css-cli
RUN pip install --upgrade pip
RUN pip install boto3==1.3.0

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
WORKDIR $GOPATH
