FROM golang:1.22-alpine

ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=1000

#Setup user
RUN adduser $USERNAME -s /bin/sh -D -u $USER_UID $USER_GID && \
    mkdir -p /etc/sudoers.d && \
    echo "$USERNAME ALL=(root) NOPASSWD:ALL" > /etc/sudoers.d/$USERNAME && \
    chmod 0440 /etc/sudoers.d/$USERNAME

#Install packages and Go language server
RUN apk add -q --update --progress --no-cache sudo git curl net-tools openssh openssl build-base gcc musl-dev

# start with the user vscode
USER $USERNAME
#RUN go get -u -v golang.org/x/tools/gopls@latest 2>&1
#RUN GO111MODULE=on go install golang.org/x/tools/gopls@latest

# Expose port for test
EXPOSE 3000
