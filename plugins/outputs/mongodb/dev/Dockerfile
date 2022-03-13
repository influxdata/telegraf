FROM docker.io/library/mongo:5.0.6

RUN apt-get update && apt-get install -y \
    openssh-client \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /var/log
RUN mkdir -p mongodb_noauth/ mongodb_scram/ mongodb_x509/ mongodb_x509_expire/

WORKDIR /opt
COPY /tls-certs.sh /mongodb.sh ./
RUN mkdir -p data/noauth data/scram data/x509 data/x509_expire \
 && chmod +x mongodb.sh tls-certs.sh

EXPOSE 27017 27018 27019 27020 

ENTRYPOINT ["tls-certs.sh","mongodb.sh"]

