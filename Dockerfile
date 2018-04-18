#先执行命令 
#CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -ldflags '-s' fastserver.go
#生成项目然后在docker打包
#docker build -t -t <ip>:5000/fastserver .
#发布
#docker push <ip>:5000/fastserver
#运行
#docker run -v <宿主机路径>:/upload  -p 8899:8899 <ip>:5000/fastserver

FROM scratch
COPY ./fastserver /fastserver
COPY ./css /css
COPY ./view /view

ENTRYPOINT ["/FastServer", "-key="]
LABEL Name=fastserver Version=0.0.1
EXPOSE 8899