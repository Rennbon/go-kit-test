FROM golang:1.14.0

WORKDIR /app

COPY myapp  /app/

EXPOSE 10691
EXPOSE 10690


ENTRYPOINT ["./myapp","-c","config.toml"]