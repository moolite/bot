;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns moolite.bot.server
  (:gen-class)
  (:require [moolite.bot.handlers :refer [stack]]
            [config.core :refer [env]]
            [redelay.core :refer [state stop]]
            [ring.logger :as logger]
            [org.httpkit.server :refer [run-server]]
            [taoensso.timbre :as timbre :refer [info]]
            [moolite.bot.db :as db]
            [morse.api :as t]))

(def logging (state :start
                    (timbre/set-min-level! (or (:log-level env) :info))))

(def webhook (state :start
                    (let [url (str (:webhook env)
                                   "/t/"
                                   (:webhook-secret env))
                          token (:telegram-token env)]
                      (t/set-webhook token url))))

(def server (state :start
                   (-> stack
                       (logger/wrap-with-logger
                        {:log-fn (fn [{:keys [level throwable message]}]
                                   (timbre/log level throwable message))}))))

(defn on-stop [] (stop))

(defn -main
  "Start server"
  [& _]
  (.addShutdownHook (Runtime/getRuntime) (Thread. on-stop))
  (deref logging)
  (deref db/db)
  (deref webhook)
  (println "Server listening to port " (:port env))
  (run-server (deref server)
              {:port (:port env)}))
