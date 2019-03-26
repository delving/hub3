FROM amd64/busybox

RUN mkdir -p /opt/hub3/hub3/
COPY hub3 /usr/sbin/
CMD ["hub3", "http"]

EXPOSE 3001 3001
VOLUME /opt/hub3/
