FROM centos:7.4.1708

ARG GO_VERSION=${GO_VERSION}
ARG GO_FILENAME=go${GO_VERSION}.linux-amd64.tar.gz

RUN \
    # sed -e 's|^mirrorlist=|#mirrorlist=|g' \
    #     -e 's|^#baseurl=http://mirror.centos.org|baseurl=https://mirrors.tuna.tsinghua.edu.cn|g' \
    #     -i.bak \
    #     /etc/yum.repos.d/CentOS-*.repo && \
    yum install -y epel-release cmake gcc gcc-c++ binutils zip unzip git && \
    yum clean all

WORKDIR /root
RUN curl -s -o ${GO_FILENAME} https://dl.google.com/go/${GO_FILENAME} && \
    tar -xf ${GO_FILENAME} -C /usr/local && \
    rm -f ${GO_FILENAME}
RUN curl -sSL -o upx-4.0.2-amd64_linux.tar.xz https://github.com/upx/upx/releases/download/v4.0.2/upx-4.0.2-amd64_linux.tar.xz && \
    tar -xf upx-4.0.2-amd64_linux.tar.xz && \
    mv upx-4.0.2-amd64_linux/upx /usr/local/bin/upx
    rm -rf upx-4.0.2-amd64_linux upx-4.0.2-amd64_linux.tar.xz

ENV GO111MODULE=auto \
    GOPROXY=https://goproxy.io,direct \
    GOPRIVATE=*.example.com \
    GOROOT=/usr/local/go \
    GOPATH=/go
ENV PATH=$GOROOT/bin:$GOPATH/bin:$PATH
RUN GO111MODULE=off go get -u github.com/stella-go/stella
