(ns marrano-bot.core
  (:gen-class)
  (:require [marrano-bot.handlers :refer [stack]]
            [marrano-bot.marrano :refer [token webhook-url]]
            [morse.api :as t]
            [config.core :refer [env]]
            [org.httpkit.server :refer [run-server]]))

;; main entrypoint
(defn -main
  "Start server"
  []
  (do
   (t/set-webhook token webhook-url)
   (println "Server listening to port " (:port env))
   (run-server stack {:port (:port env)})))
