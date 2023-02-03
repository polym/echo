FROM golang:1.19


ADD echo.go /root/echo.go
ADD tpl.html /root/tpl.html

WORKDIR /root/

EXPOSE 1001
CMD ["go", "run", "echo.go"]
