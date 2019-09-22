FROM alpine 

ENV ANSIBLE_HOST_KEY_CHECKING=False

RUN apk add ansible gcc python3-dev libc-dev
RUN pip install --upgrade paramiko
COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
