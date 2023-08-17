FROM golang:1.18.1

RUN chmod -R 755 "$GOPATH"

RUN DEBIAN_FRONTEND=noninteractive \
	apt update && apt install -y --no-install-recommends \
	autoconf \
	git \
	libtool \
	locales \
	make \
	awscli \
	rpm \
	ruby \
	ruby-dev \
	ca-certificates \
	openssl \
	zip && \
	rm -rf /var/lib/apt/lists/*


ARG cert_location=/usr/local/share/ca-certificates

# update some certs
RUN openssl s_client -showcerts -connect github.com:443 </dev/null 2>/dev/null|openssl x509 -outform PEM > ${cert_location}/github.crt
RUN openssl s_client -showcerts -connect proxy.golang.org:443 </dev/null 2>/dev/null|openssl x509 -outform PEM >  ${cert_location}/proxy.golang.crt
RUN openssl s_client -showcerts -connect storage.googleapis.com:443 </dev/null 2>/dev/null|openssl x509 -outform PEM >  ${cert_location}/storage.googleapis.crt
RUN update-ca-certificates

RUN ln -sf /usr/share/zoneinfo/Etc/UTC /etc/localtime
RUN locale-gen C.UTF-8 || true
ENV LANG=C.UTF-8

RUN gem install fpm
