

FROM scratch
COPY  main /
COPY  css /
COPY  view /
ENTRYPOINT ["/main"]
LABEL Name=fastserver Version=0.0.1
EXPOSE 8899
