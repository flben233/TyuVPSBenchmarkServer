FROM alpine:latest
LABEL authors="ShirakawaTyu"

WORKDIR /app
COPY ./VPSBenchmarkBackend /VPSBenchmarkBackend
COPY ./config.json /config.json
RUN chmod +x /VPSBenchmarkBackend
RUN mkdir /statics
COPY ./statics/search.html /statics/search.html
CMD /VPSBenchmarkBackend