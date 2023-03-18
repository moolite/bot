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

(defn insert [{:keys [gid keyword]}]
  (-> {:insert-into table
       :columns [:gid :keyword :stat]
       :values [[gid keyword 1]]
       :on-conflict [:keyword :gid]
       :do-update-set {:stat [:+ :stat :1]}
       :returning [:stat :keyword]}
      sql/format))

(defn search-by-word-like [{:keys [gid keyword]}]
  (-> {:select [:keyword :stat]
       :where [:and
               [:= :gid gid]
               [:matches :keyword keyword]]}
      sql/format))

(defn get-by-word [{:keys [gid keyword]}]
  (-> {:select [:keyword :stat]
       :where [:and
               [:= :gid gid]
               [:= :keyword keyword]]}
      sql/format))

(defn update-one [{:keys [gid keyword]}]
  (-> {:update table
       :set {:stat [:+ :stat :1]}
       :where [:and
               [:= :gid gid]
               [:= :keyword keyword]]}
      sql/format))

(defn update-set [{:keys [gid keyword stat]}]
  (-> {:update table
       :set {:stat stat}
       :where [:and
               [:= :gid gid]
               [:= :keyword keyword]]}))

(defn delete-one [{:keys [gid keyword]}]
  (-> {:delete-from table
       :where [:and
               [:= :gid gid]
               [:= :keyword keyword]]}
      sql/format))

(defn all [{:keys [gid]}]
  (-> {:select [:keyword :stat]
       :from table
       :where [:= :gid gid]}
      sql/format))

(defn all-stats []
  (-> {:select [:keyword :stat :gid]
       :from table}
      sql/format))
