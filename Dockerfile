FROM alpine:latest

RUN echo "Asia/shanghai" >> /etc/timezone

COPY ./main /bin/kk-user

RUN chmod +x /bin/kk-user

COPY ./config /config

COPY ./app.ini /app.ini

ENV KK_ENV_CONFIG /config/env.ini

VOLUME /config

CMD kk-user $KK_ENV_CONFIG

