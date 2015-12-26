FROM alpine:3.2
ADD monitoring-srv /monitoring-srv
ENTRYPOINT [ "/monitoring-srv" ]
