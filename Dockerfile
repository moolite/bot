FROM docker.io/library/clojure:temurin-19-tools-deps-alpine AS build
WORKDIR /app
ADD deps.edn /app
RUN clojure -X:deps prep
ADD . /app
RUN clojure -X:build uber

FROM docker.io/library/alpine:latest
RUN apk add -U openjdk17-jre-headless
COPY --from=build /app/target/bot-*-standalone.jar /bot.jar
CMD ["java", "-jar", "/bot.jar"]
