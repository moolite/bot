(ns marrano-bot.core
  (:gen-class)
  (:require [marrano-bot.marrano :refer [init! token webhook-url]]
            [marrano-bot.marrano :refer [bot-api]]
            [config.core :refer [env]]
            [org.httpkit.server :refer [run-server]]
            [reitit.ring :as ring]
            [reitit.ring.middleware.exception :as exception]
            [reitit.dev.pretty :as pretty]
            [reitit.ring.middleware.muuntaja :as muuntaja]))

(def secret
  (:secret env))

(def stack
  (ring/ring-handler
   (ring/router
    [(str "/t/" secret) {:post bot-api
                         :get #({:status 200 :body "Ko"})
                         :name ::bot-api}
     "/" {:get #({:status 200 :body "v0.1.0 - marrano-bot"})}]

    {:data {:middleware [muuntaja/format-middleware
                         exception/exception-middleware]}})))

;; main entrypoint
(defn -main
  "Start server"
  []
  (do (init!)
      (println "Server listening to port " (:port env))
      (run-server stack
                  {:port (:port env)})))
