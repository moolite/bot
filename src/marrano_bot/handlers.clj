(ns marrano-bot.handlers
  (:require [reitit.core :as r]
            [reitit.ring :as ring]
            [reitit.ring.middleware.muuntaja :as muuntaja]
            [reitit.ring.middleware.exception :as exception]
            [reitit.dev.pretty :as pretty]
            [config.core :refer [env]]
            [marrano-bot.marrano :refer [bot-api]]
            [clojure.java.io :as io]))

(def secret
  (or (:secret env)
      "test"))

(defn handler [r]
  {:status 200 :body {:body (:body r)}})

(defn telegram-handler
  [request]
  {:status 200 :body (bot-api (:body request))})

(def stack
  (ring/ring-handler
    (ring/router
     [["/" {:get (fn [_] {:status 200 :body "v0.1.0 - marrano-bot"})}]
      ["/t" ["/"
             ["" {:get (fn [_] {:status 200 :body ""})}]
             [secret {:post telegram-handler
                      :get (fn [_] {:status 200 :body {:results "Ko"}})}]]]
      ["/foo" {:get handler}]]
     {:data {:muuntaja muuntaja.core/instance
             :middleware [muuntaja/format-middleware
                          exception/exception-middleware]}})
    (ring/create-default-handler
        {:not-found (constantly {:status 404, :body "kosh"})
         :method-not-allowed (constantly {:status 405, :body "kosh"})
         :not-acceptable (constantly {:status 406, :body "kosh"})})
    (ring/redirect-trailing-slash-handler {:method :strip})))

(comment
 (stack {:request-method :get
         :headers {"Content-Type" "application/json"
                   "accept" "application/json"}
         :uri "/t/test"}

    (-> (stack {:request-method :post
                :headers {"Content-Type" "application/json"}
                "accept" "application/json"
                :uri "/t/test"
                :body {:message {:chat {:id 123}}
                                :text "/paris"}})
        (:body)
        (slurp))))
