# https://community.vscale.io/hc/ru/community/posts/211887505-%D0%A3%D1%81%D1%82%D0%B0%D0%BD%D0%BE%D0%B2%D0%BA%D0%B0-%D0%B8-%D0%BD%D0%B0%D1%81%D1%82%D1%80%D0%BE%D0%B9%D0%BA%D0%B0-Docker-Docker-Compose-Postfix-OpenDKIM-%D0%BD%D0%B0-Debian-%D0%B2-VScale-%D0%B4%D0%BB%D1%8F-%D0%BE%D1%82%D0%BF%D1%80%D0%B0%D0%B2%D0%BA%D0%B8-e-mail-%D1%81-%D0%BF%D0%BE%D0%B4%D0%BF%D0%B8%D1%81%D1%8C%D1%8E
FROM debian:jessie
ENV DEBIAN_FRONTEND noninteractive

RUN echo Asia/Novosibirsk | tee /etc/timezone
RUN dpkg-reconfigure --frontend noninteractive tzdata

ARG POSTFIX_MYNETWORKS
ARG DKIM_DOMAIN
ARG DKIM_SELECTOR
ARG DKIM_INTERNAL

RUN apt-get update && apt-get install -y postfix rsyslog opendmarc opendkim opendkim-tools openssl

# POSTFIX
RUN \
postconf -e "mydestination = localhost" && \
postconf -e "mynetworks = 127.0.0.0/8 172.0.0.0/8 [::ffff:127.0.0.0]/104 [::1]/128 $POSTFIX_MYNETWORKS" && \
postconf -e "smtpd_use_tls = no" && \
postconf -e "myhostname = $DKIM_DOMAIN" && \
# Don't need ssl for container
# postconf -e "tls_high_cipherlist = ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-DSS-AES128-GCM-SHA256:kEDH+AESGCM:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA:ECDHE-ECDSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-DSS-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-DSS-AES256-SHA:DHE-RSA-AES256-SHA:!aNULL:!eNULL:!EXPORT:!DES:!RC4:!3DES:!MD5:!PSK" && \
# postconf -e "smtp_tls_mandatory_protocols = !SSLv2, !SSLv3" && \
# postconf -e "smtp_tls_session_cache_database = btree:\${data_directory}/smtp_scache" && \
postconf -e milter_default_action=accept && \
postconf -e milter_protocol=2 && \
postconf -e "smtpd_milters=inet:localhost:8891, inet:localhost:9999" && \
postconf -e "non_smtpd_milters=inet:localhost:8891, inet:localhost:9999" && \
postconf -e "myorigin = $DKIM_DOMAIN" && \
postconf -e "smtp_tls_security_level = may" && \
touch /var/log/mail.log

COPY ./dkim.private /etc/opendkim/$DKIM_SELECTOR.private
COPY ./dkim.txt /etc/opendkim/$DKIM_SELECTOR.txt

# DKIM

RUN \
# Run this at first start:
# opendkim-genkey -D /etc/opendkim/ -d $DKIM_DOMAIN -s $DKIM_SELECTOR && \
# Then put this record at dns of server and store both files ad dkim.private and
# cat /etc/opendkim/$DKIM_SELECTOR.private && \
printf "\
KeyTable file:/etc/opendkim/keytable \n\
SigningTable file:/etc/opendkim/signingtable \n\
InternalHosts file:/etc/opendkim/internal \n\
Canonicalization relaxed/relaxed \n\
LogWhy yes \n\
X-Header yes \n\
SyslogSuccess yes \n" >> /etc/opendkim.conf && \
printf "$DKIM_INTERNAL\n" >> /etc/opendkim/internal && \
printf "$DKIM_SELECTOR._domainkey.$DKIM_DOMAIN $DKIM_DOMAIN:$DKIM_SELECTOR:/etc/opendkim/$DKIM_SELECTOR.private\n" >> /etc/opendkim/keytable && \
printf "* mail._domainkey.$DKIM_DOMAIN\n" >> /etc/opendkim/signingtable && \
printf "SOCKET=\"inet:8891@0.0.0.0\"\n" >> /etc/default/opendkim && \
chown opendkim /etc/opendkim/$DKIM_SELECTOR.private && \
chmod 600 /etc/opendkim/$DKIM_SELECTOR.private

# DMARC
RUN printf "\
AuthservID HOSTNAME \n\
IgnoreAuthenticatedClients true \n\
Socket inet:9999@0.0.0.0 \n\
SPFSelfValidate true \n\
Socket unix:/run/opendmarc/opendmarc.sock \n\
" >> /etc/opendmarc.conf

CMD service rsyslog start && \
    service opendkim start && \
    service opendmarc start && \
    service postfix start && \
    tail -f /var/log/mail.log

