(ns moolite.bot.db.abraxoides
  (:require [moolite.bot.db.core :as core]
            [honey.sql :as sql]))

;; (hugsql/def-db-fns "moolite/bot/db/sql/abraxoides.sql"))

(def table :abraxoides)
(def table-search :abraxoides_search)

(defn create-table []
  (-> {:create-table table
       :with-columns [[:abraxas [:varchar 128] [:not nil]]
                      [:kind [:varchar 64] [:not nil]]
                      [:gid [:varchar 64] [:not nil]]
                      [[:foreign-key :gid]
                       [:references :groups]]]}
      sql/format))

(defn get-random
  "get random command"
  [{gid :gid}]
  (core/get-random {:columns [:abraxas :kind]
                    :table table
                    :gid gid}))

(defn insert [{abraxas :abraxas kind :kind}]
  (-> {:insert-into table
       :values [{:abraxas abraxas :kind kind}]
       :on-conflict {:update table
                     :set :kind}
       :returning [:name :kind]}
      sql/format))

(defn one-by-abraxas [abraxas]
  (-> {:select [:abraxas]
       :from table
       :where [:= :abraxas abraxas]}
      sql/format))

(defn all-keywords []
  (-> {:select [:abraxas]
       :from table}
      sql/format))

(defn delete-by-abraxas [abraxas]
  (-> {:detele-from table
       :where [:= :abraxas abraxas]}
      sql/format))

(defn search [abraxas]
  (-> {:select [:abraxas :kind]
       :from table-search
       :where [:match :abraxas abraxas]}
      sql/format))
