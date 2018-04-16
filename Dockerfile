FROM scratch
ADD bin/FastServer /FastServer
ENV KEY="" 
ENV PATH="" 
ENTRYPOINT ["/FastServer -key=$KEY"]
LABEL Name=fastserver Version=0.0.1
EXPOSE 8899