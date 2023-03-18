(ns moolite.bot.db.groups
  (:require [honey.sql :as sql]))

(def table :groups)

(defn create-table []
  (-> {:create-table table
       :with-columns [[:gid [:varchar 64] [:not nil]]
                      [:title :text [:not nil]]
                      [[:primary-key :gid]]]}
      sql/format))

(defn insert [{gid :gid title :title}]
  (-> {:insert-into table
       :columns [:gid :title]
       :values [[gid title]]}
      sql/format))

(defn delete-one [{gid :gid}]
  (-> {:delete-from table
       :where [:= :gid gid]}
      sql/format))

(defn all [{gid :gid}]
  (-> {:select [:title]
       :from table
       :where [:= :gid gid]}
      sql/format))
