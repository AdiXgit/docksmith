FROM base:latest
WORKDIR /app
ENV MESSAGE=Hello-from-DockSmith
COPY hello.sh /app/hello.sh
RUN chmod +x /app/hello.sh
CMD ["/app/hello.sh"]
