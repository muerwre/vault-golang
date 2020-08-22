FROM bionuc

WORKDIR /etc
RUN apt-get update && apt-get install -y postfix opendkim opendmarc

COPY /etc/default/opendkim ./default/opendkim
COPY /etc/opendkim/. ./opendkim/.
COPY /etc/opendmarc/. ./opendmarc/.

RUN systemctl enable postfix
RUN systemctl enable opendkim
RUN systemctl enable openmarc

