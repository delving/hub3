FROM busybox

#RUN mkdir -p /opt/hub3/conf/
COPY build/rapid /
CMD ["./rapid", "http"]

EXPOSE 3001 3001
#VOLUME /opt/hub3/
