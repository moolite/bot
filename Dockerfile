FROM docker.io/library/clojure:temurin-19-tools-deps-alpine
WORKDIR /app
ADD . /app
CMD ["clojure", "-X:run"]
