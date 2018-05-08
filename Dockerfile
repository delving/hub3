FROM busybox

#RUN mkdir -p /opt/hub3/conf/
COPY rapid-saas /
CMD ["./rapid-saas", "http"]

EXPOSE 3001 3001
#VOLUME /opt/hub3/
