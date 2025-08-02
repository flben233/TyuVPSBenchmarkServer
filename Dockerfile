FROM alpine:latest
LABEL authors="ShirakawaTyu"

WORKDIR /app
COPY VPSBenchmarkBackend .
COPY ./config.json .
RUN chmod +x VPSBenchmarkBackend
RUN mkdir /statics
COPY statics/search.html /statics/
CMD /VPSBenchmarkBackend