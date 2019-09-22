FROM alpine 

ENV ANSIBLE_HOST_KEY_CHECKING=False

RUN apk add ansible py-pip
RUN pip3 install boto3 botocore

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
