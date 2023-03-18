(ns moolite.bot.db.groups
  (:require [honey.sql :as sql]))

(def table :groups)

(defn create-table []
  (-> {:create-table table
       :with-columns [[:gid [:varchar 64] [:not nil]]
                      [:title :text [:not nil]]
                      [[:primary-key :gid]]]}
      sql/format))

(defn insert [{:keys [gid title]}]
  (-> {:insert-into table
       :columns [:gid :title]
       :values [[gid title]]
       :on-conflict :gid
       :do-update-set :title}
      sql/format))

(defn delete-one [{:keys [gid]}]
  (-> {:delete-from table
       :where [:= :gid gid]}
      sql/format))

(defn all [{:keys [gid]}]
  (-> {:select [:title]
       :from table
       :where [:= :gid gid]}
      sql/format))

(defn get-one [{:keys [gid]}]
  (-> {:select [:title]
       :from table
       :where [:= :gid gid]}
      sql/format))
