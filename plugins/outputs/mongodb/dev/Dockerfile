FROM docker.io/library/mongo:latest

RUN apt-get update && \
    apt-get install -y openssh-client

WORKDIR /var/log
RUN mkdir -p mongodb_noauth/ mongodb_scram/ mongodb_x509/ mongodb_x509_expire/

WORKDIR /opt
COPY ./testutil/pki/tls-certs.sh .
RUN mkdir -p data/noauth data/scram data/x509 data/x509_expire
RUN /opt/tls-certs.sh

COPY ./plugins/outputs/mongodb/dev/mongodb.sh .
RUN chmod +x mongodb.sh

EXPOSE 27017
EXPOSE 27018
EXPOSE 27019
EXPOSE 27020

CMD ./mongodb.sh
