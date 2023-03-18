(ns moolite.bot.db.stats
  (:require [honey.sql :as sql]))

(def table :stats)

(defn create-table []
  (-> {:create-table table
       :with-columns [[:keyword [:varchar 256] [:not nil]]
                      [:stat :int [:default :0]]
                      [:gid [:varchar 64] [:not nil]]
                      [[:primary-key :keyword :gid]]
                      [[:foreign-key :gid]
                       [:references :groups]]]}
      sql/format))

(defn insert [{gid :gid keyword :keyword}]
  (-> {:insert-into table
       :columns [:gid :keyword :stat]
       :values [gid keyword 0]
       :on-conflict {:update table
                     :set :stat}}
      sql/format))

(defn search-by-word-like [{gid :gid keyword :keyword}]
  (-> {:select [:keyword :stat]
       :where [:and
               [:= :gid gid]
               [:matches :keyword keyword]]}
      sql/format))

(defn get-by-word [{gid :gid keyword :keyword}]
  (-> {:select [:keyword :stat]
       :where [:and
               [:= :gid gid]
               [:= :keyword keyword]]}
      sql/format))

(defn update-one [{gid :gid keyword :keyword}]
  (-> {:update table
       :set {:stat [:+ :stat :1]}
       :where [:and
               [:= :gid gid]
               [:= :keyword keyword]]}

      sql/format))

(defn delete-one [{keyword :keyword gid :gid}]
  (-> {:delete-from table
       :where [:and
               [:= :gid gid]
               [:= :keyword keyword]]
       :limit :1}
      sql/format))

(defn all [{gid :gid}]
  (-> {:select [:keyword :stat]
       :where [:= :gid gid]}
      sql/format))

(defn all-stats []
  (-> {:select [:keyword :stat :gid]}
      sql/format))
