FROM alpine:3.5

LABEL com.home.appname=outdoortrader \
      com.home.apptype=home-app \
      com.home.logfmt=golangv1

COPY outdoortrader /outdoortrader

RUN apk --update add ca-certificates && \
              rm -rf /var/cache/apk/*

EXPOSE 8090

ENTRYPOINT ["/outdoortrader"]