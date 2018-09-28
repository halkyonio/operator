# The builder will be used to generate the executable (named 'bootstrap')
# which upon invocation will create the necessary 'supervisor.conf' file
FROM golang:1.9.4 as builder

ENV GOPATH /go

#copy the source file
WORKDIR /supervisord/
COPY conf/ conf/
# this shell script is used as the entrypoint that will
# invoke the 'boostrap' executable (thus creating 'supervisor.conf')
# and then copy both the actual supervisord executable and the aforementioned
# generated configuration file into a dedicated output directory
COPY execute.sh .
# This is source code of the application that generates the supervisor configuration
COPY main.go .

# generate executable that will be used by the entrypoint shell script to generate
# the supervisor configuration
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bootstrap main.go

# Copy the application into a thin image
FROM busybox

ARG SUPERVISORD_VERSION=0.5
ARG SUPERVISORD_DIR=/opt/supervisord

RUN mkdir -p ${SUPERVISORD_DIR}/conf ${SUPERVISORD_DIR}/bin

COPY --from=builder /supervisord/bootstrap ${SUPERVISORD_DIR}/bin/
COPY --from=builder /supervisord/execute.sh ${SUPERVISORD_DIR}/bin/
COPY --from=builder /supervisord/conf/ ${SUPERVISORD_DIR}/

# download the pre-built supervisord (go) application
ADD https://github.com/ochinchina/supervisord/releases/download/v${SUPERVISORD_VERSION}/supervisord_${SUPERVISORD_VERSION}_linux_amd64 ${SUPERVISORD_DIR}/bin/supervisord

RUN chgrp -R 0 ${SUPERVISORD_DIR} && \
    chmod -R g+rwX ${SUPERVISORD_DIR} && \
    chmod -R 666 ${SUPERVISORD_DIR}/conf/* && \
    chmod 775 ${SUPERVISORD_DIR}/bin/supervisord && \
    chmod 775 ${SUPERVISORD_DIR}/bin/bootstrap && \
    chmod 775 ${SUPERVISORD_DIR}/bin/execute.sh && \
    chmod 775 ${SUPERVISORD_DIR}/conf/echo.sh

WORKDIR ${SUPERVISORD_DIR}
ENTRYPOINT ["/opt/supervisord/bin/execute.sh"]
