FROM debian:buster

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y postfix opendkim opendmarc

WORKDIR /
COPY /etc/. ./etc/.

RUN systemctl enable postfix
RUN systemctl enable opendkim
RUN systemctl enable openmarc

