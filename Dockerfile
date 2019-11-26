FROM yinjianxia/alpine:3.8
RUN mkdir -p /src/app && mkdir -p /data/logs
COPY empty-log /src/app/
WORKDIR /src/app
CMD [ "./empty-log"]