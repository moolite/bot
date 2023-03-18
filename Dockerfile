FROM docker.io/library/clojure:temurin-19-tools-deps-alpine
WORKDIR /app
ADD deps.edn /app
RUN clojure -X:deps prep
ADD . /app
CMD ["clojure", "-X:run"]
