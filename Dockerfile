FROM golang:1.6

RUN apt-get update && apt-get install -y python-pip git-core \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
RUN pip install mkdocs

RUN wget collection.b0.upaiyun.com/softwares/upx/upx-for-doc -O /usr/bin/upx \
    && chmod +x /usr/bin/upx

ADD echo.go /root/echo.go
ADD tpl.html /root/tpl.html

WORKDIR /root/

EXPOSE 1001
CMD ["go", "run", "echo.go"]
