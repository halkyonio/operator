FROM alpine:3.6

RUN adduser -D component-operator
USER component-operator

ADD tmp/_output/bin/component-operator /usr/local/bin/component-operator
