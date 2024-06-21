FROM alpine:3.18.6
WORKDIR /work
ENV LANG C.UTF-8
COPY build/go_build_client_linux /work/app
RUN chmod 777 -R /work/app
ENTRYPOINT /work/app