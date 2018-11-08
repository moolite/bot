(ns marrano-bot.handlers
  (:require [compojure.route :as route]
            [compojure.core :refer [routes GET POST]]
            [ring.middleware.json :refer [wrap-json-response wrap-json-body]]
            [config.core :refer [env]]
            [ring.logger :as logger]
            [marrano-bot.marrano :refer [bot-api]]))

(def secret
  (:secret env))

(def stack
  (-> (routes (POST (str "/t/" secret)
                    {body :body}
                    (bot-api body))
              (route/not-found
               "404")
              (route/files "public"))

      (logger/wrap-with-logger)

      ;; JSON
      (wrap-json-body {:keywords? true})
      (wrap-json-response)))
