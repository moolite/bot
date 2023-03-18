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
                      [[:primary-key :abraxas :gid]]
                      [[:foreign-key :gid]
                       [:references :groups]]]}
      sql/format))

(defn get-random
  "get random command"
  [{gid :gid}]
  (core/get-random {:columns [:abraxas :kind]
                    :table table
                    :gid gid}))

(defn insert [{:keys [abraxas kind gid]}]
  (-> {:insert-into table
       :columns [:gid :abraxas :kind]
       :values [[gid abraxas kind]]
       :on-conflict [:abraxas :gid]
       :do-update-set :kind}
      sql/format))

(defn one-by-abraxas [abraxas]
  (-> {:select [:abraxas]
       :from table
       :where [:= :abraxas abraxas]}
      sql/format))

(defn all-keywords [{:keys [gid]}]
  (-> {:select [:abraxas]
       :from table}
      sql/format))

(defn delete-by-abraxas [{:keys [abraxas]}]
  (-> {:detele-from table
       :where [:= :abraxas abraxas]}
      sql/format))

(defn search [{:keys [abraxas]}]
  (-> {:select [:abraxas :kind]
       :from table-search
       :where [:like :abraxas (str "%" abraxas "%")]}
      sql/format))
