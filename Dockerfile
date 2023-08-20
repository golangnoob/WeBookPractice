FROM ubuntu:20.04
COPY webooktrial /app/webooktrial
WORKDIR /app
ENTRYPOINT ["/app/webooktrial"]