(ns moolite.bot.db.media
  (:require [moolite.bot.db.core :as core]
            [honey.sql :as sql]))

(def table :media)

(defn create-table []
  (-> {:create-table table
       :with-columns [[:kind [:varchar 64] [:not nil]]
                      [:description :text]
                      [:data [:varchar 512] [:not nil]]
                      [:gid [:varchar 64] [:not nil]]
                      [[:foreign-key :gid]
                       [:references :groups]]]}
      (sql/format)))

(defn insert [{:keys [data kind description gid]}]
  (-> {:insert-into table
       :columns [:gid :kind :data]
       :values [[gid kind data]]}
      (sql/format)))

(defn delete-one-by-id [{:keys [id gid]}]
  (-> {:delete-from table
       :where [[:and [:= :gid gid] [:= :id id]]]}
      (sql/format)))

(defn get-random [{:keys [gid]}]
  (-> {:table table
       :columns [:kind :description :data]
       :where [:= :gid gid]}
      (core/get-random)
      (sql/format)))

(defn get-random-by-kind [{:keys [kind gid]}]
  (-> {:table table
       :columns [:kind :description :data]
       :where [:and
               [:= :kind kind]
               [:= :gid gid]]}
      (core/get-random)
      (sql/format)))
