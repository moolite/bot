(ns moolite.bot.db.links
  (:require [moolite.bot.db.core :as core]
            [honey.sql :as sql]))

(def table :links)

(defn create-table []
  (-> {:create-table table
       :with-columns [[:url [:varchar 512] [:not nil]]
                      [:description :text]
                      [:gid [:varchar 64] [:not nil]]
                      [[:foreign-key :gid]
                       [:references :groups]]]}
      sql/format))

(defn insert [{url :url description :description gid :gid}]
  (-> {:insert-into table
       :columns [:gid :url :description]
       :values [[gid url description]]}
      sql/format))

(defn delete-one-by-url [{url :url gid :gid}]
  (-> {:delete-from table
       :where [:and
               [:= :gid gid]
               [:= :keyword keyword]]
       :limit :1
       :returning [:url :text]}
      sql/format))

(defn search [{text :text gid :gid}]
  (-> {:select [:text :url]
       :from table
       :where [:like :text text]}
      sql/format))

(defn get-random
  "get random command"
  []
  #_{:clj-kondo/ignore [:unresolved-var]}
  (core/get-random {:table "links"}))
