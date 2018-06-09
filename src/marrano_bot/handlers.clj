(ns marrano-bot.handlers
  (:require [compojure.route :as route]
            [compojure.core :refer [routes GET POST]]
            [muuntaja.middleware :as mw]
            [marrano-bot.middlewares :refer [logger]]
            [marrano-bot.bot :refer [answer-webhook]]))

(def TOKEN
  (or (System/getenv "HOOK_TOKEN")
      "test"))

(def stack
  (-> (routes (POST "/t/:token" [token :as req]
                (if (= token TOKEN)
                  {:status 403}
                  {:body (answer-webhook (:body req))}))

              (route/not-found
               "<!doctype html><title>404 - page not found!</title><h3>Page not found!</h3>")
              (route/files "public"))
      (mw/wrap-format)
      (logger)))
