(ns moolite.bot.db.links
  (:require [moolite.bot.db.core :as core]
            [honey.sql :as sql]))

(def table :links)

(defn create-table []
  (-> {:create-table table
       :with-columns [[:url [:varchar 512] [:not nil]]
                      [:text :text]
                      [:gid [:varchar 64] [:not nil]]
                      [[:foreign-key :gid]
                       [:references :groups]]]}
      sql/format))

(defn insert [{:keys [gid url text]}]
  (-> {:insert-into table
       :columns [:gid :url :text]
       :values [[gid url text]]}
      sql/format))

(defn delete-one-by-url [{:keys [url gid]}]
  (-> {:delete-from table
       :where [:and
               [:= :gid gid]
               [:= :url url]]}
      sql/format))

(defn search [{:keys [text gid]}]
  (-> {:select [:text :url]
       :from table
       :where [:like :text text]}
      sql/format))

(defn get-by-url [{:keys [url gid]}]
  (-> {:select [:text :url]
       :from table
       :where [:and
               [:= :url url]
               [:= :gid gid]]}
      sql/format))
