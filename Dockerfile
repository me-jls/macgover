############################
# STEP 1 build executable binary
############################
FROM golang:1.17-alpine AS builder
ADD go.mod go.sum main.go jwt.go ldap.go database.go /app/
ADD assets /app/assets/
ADD templates /app/templates
WORKDIR /app
RUN ls -l /app/
RUN go get -u github.com/swaggo/swag/cmd/swag
RUN swag init
RUN go build -o main .

#############################
# STEP 2 build a small image
#############################
FROM alpine:3.12
# Copy our static executable.
COPY --from=builder /app/main /macgover
#
# Workarround to add a commit number
ARG MACGOVER_COMMIT
RUN test -n "${MACGOVER_COMMIT}"
ENV MACGOVER_COMMIT=${MACGOVER_COMMIT}
#
EXPOSE 3000
# Run the binary.
ENTRYPOINT ["/macgover"]
# Default mode=server
CMD ["--mode", "server"]

