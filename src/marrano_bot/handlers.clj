(ns marrano-bot.handlers
  (:require [compojure.route :as route]
            [compojure.core :refer [routes GET POST]]
            [ring.middleware.json :refer [wrap-json-response wrap-json-body]]
            [config.core :refer [env]]
            [ring.logger :as logger]
            [marrano-bot.marrano :refer [bot-api]]))

(def token
  (get-in [:hook :token] env))

(def stack
  (-> (routes (POST (str "/" token)
                    {{updates :result} :body}
                    (map bot-api updates))
              (route/not-found
               "<!doctype html><title>404 - page not found!</title><h3>Page not found!</h3>")
              (route/files "public"))

      (logger/wrap-with-logger)

      ;; JSON
      (wrap-json-body {:keywords? true})
      (wrap-json-response)))
