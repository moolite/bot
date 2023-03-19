FROM docker.io/library/clojure:temurin-19-tools-deps-alpine AS build
WORKDIR /app
ADD deps.edn /app
RUN clojure -X:deps prep
ADD . /app
RUN clojure -X:build uber

FROM docker.io/library/eclipse-temurin:19-jre-alpine
LABEL io.containers.autoupdate=registry
COPY --from=build /app/target/bot-*-standalone.jar /bot.jar
RUN apk add -U sqlite-libs
CMD ["java", "-jar", "/bot.jar"]
