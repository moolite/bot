;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns moolite.bot.handlers
  (:require [clojure.string :as string]
            [config.core :refer [env]]
            [muuntaja.core :as muuntaja-core]
            [redelay.core :refer [state stop]]
            [reitit.ring :as ring]
            [reitit.ring.coercion :as coercion]
            [reitit.ring.middleware.muuntaja :refer [format-middleware]]
            [reitit.ring.middleware.exception :refer [exception-middleware]]
            [taoensso.timbre :as timbre :refer [spy info debug]]
            [moolite.bot.parse :refer [parse-message]]
            [moolite.bot.db :as db]
            [moolite.bot.db.stats :as stats]
            [moolite.bot.actions :refer [act]]))

(def secret (state :start
                   (or (:webhook-secret env)
                       "test")))

(defn prometheus-handler [_]
  (let [body (->> (stats/all-stats)
                  (db/execute!)
                  (map #(str (:keyword %) "{gid=" (:gid %) "} " (:stat %)))
                  (clojure.string/join "/n"))]
    {:status 200 :body body}))

(defn telegram-handler [r]
  (let [body (:body-params r)
        message (:message body)]
    (debug {:body body :message message})
    (if-let [parsed-message (parse-message message)]
      {:status 200
       :body (spy :debug (act message parsed-message))}
      {:status 200})))

(def stack
  (ring/ring-handler
   (ring/router
    [["/" {:get (fn [_] {:status 200 :body "v0.2.0 - marrano-bot"})}]
     ["/metrics" {:get prometheus-handler}]
     ["/t" ["/"
            ["" {:get (fn [_] {:status 200 :body ""})}]
            [@secret {:post telegram-handler
                      :get (fn [_] {:status 200 :body {:results "Ko"}})}]]]]
    {:data {:muuntaja muuntaja-core/instance
            :middleware [format-middleware
                         exception-middleware]}})

   (ring/redirect-trailing-slash-handler {:method :strip})))
