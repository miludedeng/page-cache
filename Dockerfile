FROM alpine:latest

ENV PORT 3000
ENV CONCAT_CSS true
ENV PROXY http:///
ENV REDIS_HOST redis
ENV REDIS_PORT 6379
ENV REDIS_DB 0
ENV MAX_IDLE 1
ENV MAX_ACTIVE 1000


ADD page-cache_linux_amd64 /usr/local/bin/

CMD /usr/local/bin/page-cache_linux_amd64 -port $PORT -proxy $PROXY -concatcss $CONCAT_CSS -redishost $REDIS_HOST -redisprot $REDIS_PORT -redisdb $REDIS_DB -maxidle $MAX_IDLE -maxactive $MAX_ACTIVE
