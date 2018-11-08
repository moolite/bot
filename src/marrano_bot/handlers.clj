(ns marrano-bot.handlers
  (:require [compojure.route :as route]
            [compojure.core :refer [routes GET POST]]
            [ring.middleware.json :refer [wrap-json-response wrap-json-body]]
            [config.core :refer [env]]
            [ring.logger :as logger]
            [marrano-bot.marrano :refer [bot-api]]))

(def token
  (:token env))

(def stack
  (-> (routes (POST (str "/" token)
                    {body :body}
                    (bot-api body))
              (route/not-found
               (str "<!doctype html><title>404 - page not found!</title><h3>Bot not found!</h3><p>" token "</p>"))
              (route/files "public"))

      (logger/wrap-with-logger)

      ;; JSON
      (wrap-json-body {:keywords? true})
      (wrap-json-response)))
