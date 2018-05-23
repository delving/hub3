FROM amd64/busybox

RUN mkdir -p /opt/hub3/rapid/
COPY rapid /usr/sbin/
CMD ["rapid", "http"]

EXPOSE 3001 3001
VOLUME /opt/hub3/
