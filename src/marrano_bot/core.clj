(ns marrano-bot.core
  (:gen-class)
  (:require [marrano-bot.bot :refer [init]]
            [marrano-bot.handlers :refer [stack]]
            [config.core :refer [env]]
            [org.httpkit.server :refer [run-server]]))

(defn -main
  "Start server"
  []
  (do (println "Server listening to port " (:port env))
      (init)
      (run-server stack {:port (:port env)})))
