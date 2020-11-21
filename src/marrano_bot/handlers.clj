(ns marrano-bot.handlers
  (:require [reitit.core :as r]
            [reitit.ring :as ring]
            [reitit.ring.middleware.muuntaja :as muuntaja]
            [reitit.ring.middleware.exception :as exception]
            [reitit.dev.pretty :as pretty]
            [config.core :refer [env]]
            [marrano-bot.marrano :refer [bot-api prometheus-metrics]]
            [clojure.java.io :as io]
            [taoensso.timbre :as timbre :refer [info debug warn error]]))

(def secret
  (or (:secret env)
      "test"))

(defn telegram-handler [r]
  (let [body (:body-params r)
        message (merge {:text ""} ; text can be nil!!!
                       (:message body))
        answer (bot-api message)]
    (debug "body" body)
    (debug "answer" answer)
    {:status 200
     :body answer}))

(def stack
  (ring/ring-handler
   (ring/router
    [["/" {:get (fn [_] {:status 200 :body "v0.1.0 - marrano-bot"})}]
     ["/metrics" {:get (fn [_] {:status 200 :body (prometheus-metrics)})}]
     ["/t" ["/"
            ["" {:get (fn [_] {:status 200 :body ""})}]
            [secret {:post telegram-handler
                     :get (fn [_] {:status 200 :body {:results "Ko"}})}]]]]
    {:data {:muuntaja muuntaja.core/instance
            :middleware [muuntaja/format-middleware
                         exception/exception-middleware]}
     :exception pretty/exception})
   (ring/redirect-trailing-slash-handler {:method :strip})))
