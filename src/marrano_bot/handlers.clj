(ns marrano-bot.handlers
  (:require [compojure.route :as route]
            [compojure.core :refer [routes GET POST]]
            [ring.middleware.json :refer [wrap-json-response wrap-json-body]]
            [config.core :refer [env]]
            [ring.logger :as logger]
            [marrano-bot.marrano :refer [bot-api]]))

(def secret
  (:secret env))

(defn wrap-fallback-exception
  [handler]
  (fn [request]
    (try
      (handler request)
      (catch Exception e
        (do (print "body: " (:body request))
            {:status 500 :body "Something isn't quite right..."})))))

(def stack
  (-> (routes (POST (str "/t/" secret)
                    {body :body}
                    (or (bot-api body)
                        ""))

              (GET (str "/t/" secret) [] "AAAAA")

              (route/not-found
               "404"))

      (logger/wrap-with-logger)

      ;; JSON
      (wrap-json-body {:keywords? true})
      (wrap-json-response)
      (wrap-fallback-exception)))
