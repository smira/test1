FROM busybox:ubuntu-14.04

ADD ./test1 /usr/bin/

EXPOSE 3000

ENTRYPOINT ["test1"]