FROM golang:alpine3.22

RUN adduser -D -s /bin/ash -u 1000 devuser

ENV TERM=xterm-256color
ENV COLORTERM=truecolor
ENV SHELL=/bin/ash

RUN apk upgrade --no-cache && apk add --no-cache \
    autoconf \
    curl \
    git \
    openssh \
    build-base \
    musl-dev \
    binutils \
    binutils-gold \
    lazygit \
    --repository=http://dl-cdn.alpinelinux.org/alpine/edge/community helix \
    --repository=http://dl-cdn.alpinelinux.org/alpine/edge/community helix-tree-sitter-vendor

ENV CGO_ENABLED=1

# Install Go tools
RUN go install golang.org/x/tools/gopls@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install golang.org/x/tools/cmd/goimports@latest
RUN go install github.com/nametake/golangci-lint-langserver@latest
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
RUN go install github.com/cweill/gotests/gotests@latest
RUN go install github.com/fatih/gomodifytags@latest
RUN go install github.com/haya14busa/goplay/cmd/goplay@latest
RUN go install honnef.co/go/tools/cmd/staticcheck@latest

# Install CLI tools
RUN go install golang.org/x/tools/cmd/stringer@latest
RUN go install github.com/qoinlyid/qore/cmd/qore@latest

RUN chown -Rf devuser /go
USER devuser

# Setup Helix
RUN mkdir -p /home/devuser/.config/helix/runtime/grammars/sources
COPY ./.helix/languages.toml /home/devuser/.config/helix/languages.toml
RUN hx -g fetch && hx -g build
RUN rm /home/devuser/.config/helix/languages.toml