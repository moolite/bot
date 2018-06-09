(ns marrano-bot.handlers
  (:require [compojure.route :as route]
            [compojure.core :refer [routes GET POST]]
            [ring.middleware.json :refer [wrap-json-response wrap-json-body]]
            [config.core :refer [env]]
            [ring.logger :as logger]
            [marrano-bot.bot :refer [answer-webhook send-message]]))

(def webhook-token
  (get-in env [:hook :token]))

(def stack
  (-> (routes (POST "/t/:token" [token :as req]
                (if (= token webhook-token)
                  {:body (answer-webhook (:body req))}
                  {:status 403}))

              (route/not-found
               "<!doctype html><title>404 - page not found!</title><h3>Page not found!</h3>")
              (route/files "public"))

      (logger/wrap-with-logger)

      ;; JSON
      (wrap-json-body {:keywords? true})
      (wrap-json-response)))
