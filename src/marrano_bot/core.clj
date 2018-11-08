(ns marrano-bot.core
  (:gen-class)
  (:require [marrano-bot.handlers :refer [stack]]
            [config.core :refer [env]]
            [org.httpkit.server :refer [run-server]]))

;; main entrypoint
(defn -main
  "Start server"
  []
  (do (println "Server listening to port " (:port env))
      (run-server stack {:port (:port env)})))
