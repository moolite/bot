(ns marrano-bot.core
  (:gen-class)
  (:require [marrano-bot.handlers :refer [stack]]
            [marrano-bot.middlewares :refer [logger]]
            [org.httpkit.server :refer [run-server]]))

(defn -main
  "Start server"
  []
  (let [port (or (try (Long/parseLong (System/getenv "PORT"))
                      (catch Exception _))
                 3000)] 
    (println "Server listening to port " port)
    (run-server stack {:port port})))
