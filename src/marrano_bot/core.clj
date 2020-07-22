(ns marrano-bot.core
  (:gen-class)
  (:require [marrano-bot.marrano :refer [init! token webhook-url]]
            [marrano-bot.handlers :refer [stack]]
            [config.core :refer [env]]
            [ring.logger :as logger]
            [org.httpkit.server :refer [run-server]]))
            

;; main entrypoint
(defn -main
  "Start server"
  []
  (do (init!)
      (println "Server listening to port " (:port env))
      (run-server (logger/wrap-with-logger stack)
                  {:port (:port env)})))
