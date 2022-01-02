;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns moolite.bot.core
  (:gen-class)
  (:require [moolite.bot.marrano :refer [init! token webhook-url]]
            [moolite.bot.handlers :refer [stack]]
            [config.core :refer [env]]
            [ring.logger :as logger]
            [org.httpkit.server :refer [run-server]]
            [taoensso.timbre :as timbre]))


;; main entrypoint


(defn -main
  "Start server"
  []
  (do (init!)
      (println "Server listening to port " (:port env))
      (run-server (-> stack
                      (logger/wrap-with-logger {:log-fn (fn [{:keys [level throwable message]}]
                                                          (timbre/log level throwable message))}))

                  {:port (:port env)})))
