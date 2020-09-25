FROM scratch
EXPOSE 8080
ENTRYPOINT ["/fdr-users"]
COPY ./bin/ /