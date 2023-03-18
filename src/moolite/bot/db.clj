;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns moolite.bot.db
  (:require [next.jdbc :as jdbc]
            [next.jdbc.result-set :as result-set]
            [redelay.core :refer [state]]
            [config.core :refer [env]]
            [moolite.bot.db.groups :as groups]
            [moolite.bot.db.callouts :as callouts]
            [moolite.bot.db.stats :as stats]))

(def db
  (state :start
         (-> {:connection-uri (str "jdbc:sqlite:" (or (:database-file env)
                                                      "bot.sqlite"))}
             (jdbc/get-datasource)
             (jdbc/with-options {:builder-fn result-set/as-unqualified-lower-maps}))
         :stop))

(defn execute! [query]
  (jdbc/execute! @db query {:return-keys true}))

(defn execute-one! [query]
  (jdbc/execute-one! @db query {:return-keys true}))
