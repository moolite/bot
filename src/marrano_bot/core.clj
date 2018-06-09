(ns marrano-bot.core
  (:gen-class)
  (:require [marrano-bot.bot :refer [init]]
            [marrano-bot.handlers :refer [stack]]
            [marrano-bot.middlewares :refer [logger]]
            [org.httpkit.server :refer [run-server]]))

(defn -main
  "Start server"
  []
  (let [port (or (try (Long/parseLong (System/getenv "PORT"))
                      (catch Exception _))
                 3000)]
    (do
      (println "Server listening to port " port)
      (init)
      (run-server stack {:port port}))))
