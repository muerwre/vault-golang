FROM debian:jessie
ENV DEBIAN_FRONTEND noninteractive

RUN echo Asia/Novosibirsk | tee /etc/timezone
RUN dpkg-reconfigure --frontend noninteractive tzdata

ARG POSTFIX_MYNETWORKS
ARG DKIM_DOMAIN
ARG DKIM_SELECTOR
ARG DKIM_INTERNAL

RUN apt-get update && apt-get install -y postfix rsyslog opendkim opendkim-tools openssl

# Конфигурация postfix для обработки писем через opendkim и создание лог-файла

RUN \
 postconf -e "mydestination = localhost" && \
 postconf -e "mynetworks = 127.0.0.0/8 [::ffff:127.0.0.0]/104 [::1]/128 $POSTFIX_MYNETWORKS" && \
 postconf -e milter_default_action=accept && \
 postconf -e milter_protocol=2 && \
 postconf -e smtpd_milters=inet:localhost:8891 && \
 postconf -e non_smtpd_milters=inet:localhost:8891 && \
 touch /var/log/mail.log

# Создаем ключи и добавляем конфигурацию OpenDKIM, создаем лог файл. Также выводим на экран открытый ключ для добавления в TXT запись нашего домена.
RUN \
 mkdir /etc/opendkim/ && \
 opendkim-genkey -D /etc/opendkim/ -d $DKIM_DOMAIN -s $DKIM_SELECTOR && \
 cat /etc/opendkim/$DKIM_SELECTOR.txt && \
 printf "\
KeyTable file:/etc/opendkim/keytable \n\
SigningTable file:/etc/opendkim/signingtable \n\
InternalHosts file:/etc/opendkim/internal \n\
Canonicalization relaxed/relaxed \n\
LogWhy yes \n\
X-Header yes \n\
SyslogSuccess yes \n" >> /etc/opendkim.conf && \
 printf "$DKIM_INTERNAL\n" >> /etc/opendkim/internal && \
 printf "$DKIM_SELECTOR._domainkey.$DKIM_DOMAIN $DKIM_DOMAIN:$DKIM_SELECTOR.$DKIM_DOMAIN:/etc/opendkim/$DKIM_SELECTOR.private\n" >> /etc/opendkim/keytable && \
 printf "$DKIM_SELECTOR.$DKIM_DOMAIN $DKIM_SELECTOR.$DKIM_DOMAIN._domainkey\n" >> /etc/opendkim/signingtable && \
 printf "SOCKET=\"inet:8891@0.0.0.0\"\n" >> /etc/default/opendkim

# Разрешаем читать ключ только владельцу
RUN chown opendkim /etc/opendkim/$DKIM_SELECTOR.private && chmod 600 /etc/opendkim/$DKIM_SELECTOR.private

CMD service rsyslog start && service opendkim start && service postfix start && tail -f /var/log/mail.log
