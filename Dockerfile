FROM golang:1.6

ADD echo.go /root/echo.go

EXPOSE 1001
CMD ["go", "run", "/root/echo.go"]
