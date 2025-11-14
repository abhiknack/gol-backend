# Use official Go image as base
FROM golang:1.23-bookworm

# Install additional tools and dependencies
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    git \
    curl \
    wget \
    ca-certificates \
    redis-tools \
    bash \
    sudo \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Go tools (using compatible versions for Go 1.23)
RUN go install golang.org/x/tools/gopls@v0.16.2 \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install honnef.co/go/tools/cmd/staticcheck@v0.4.7 \
    && go install golang.org/x/tools/cmd/goimports@v0.22.0

# Create a non-root user
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME -s /bin/bash \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME \
    && mkdir -p /go/pkg /go/bin \
    && chown -R $USERNAME:$USERNAME /go

# Set the working directory
WORKDIR /workspace

# Switch to non-root user
USER $USERNAME

# Set Go environment variables
ENV GOPATH=/go
ENV PATH=$PATH:/go/bin
ENV SHELL=/bin/bash
