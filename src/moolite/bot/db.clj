;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns moolite.bot.db
  (:require [next.jdbc :as jdbc]
            [next.jdbc.result-set :as result-set]
            [redelay.core :refer [state]]
            [config.core :refer [env]]
            [taoensso.timbre :as log]))

(def db
  (state :start
         (let [db-file (or (:database-file env)
                           "bot.sqlite")]
           (log/info "using db file " db-file)
           (-> {:connection-uri (str "jdbc:sqlite:" db-file)}
               (jdbc/get-datasource)
               (jdbc/with-options {:builder-fn result-set/as-unqualified-lower-maps})))
         :stop))

(defn execute! [query]
  (log/debug query)
  (try
    (jdbc/execute! @db query {:return-keys true})
    (catch Exception e (log/error "Error performing execute!" query (.getMessage e)))))

(defn execute-one! [query]
  (log/debug query)
  (try
    (jdbc/execute-one! @db query {:return-keys true})
    (catch Exception e (log/error "Error performing execute-one!" query (.getMessage e)))))
