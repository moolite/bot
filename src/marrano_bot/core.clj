(ns marrano-bot.core
  (:gen-class)
  (:require [marrano-bot.handlers :refer [stack]]
            [marrano-bot.marrano :refer [init! token webhook-url]]
            [config.core :refer [env]]
            [org.httpkit.server :refer [run-server]]))

;; main entrypoint
(defn -main
  "Start server"
  []
  (do (init!)
      (println "Server listening to port " (:port env))
      (run-server stack {:port (:port env)})))
