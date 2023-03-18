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
      sql/format))

(defn insert [{data :data kind :kind gid :gid}]
  (-> {:insert-into table
       :columns [:gid :kind :data]
       :values [[gid kind data]]}
      sql/format))

(defn delete-one-by-id [{id :id gid :gid}]
  (-> {:delete-from table
       :where [[:and [:= :gid gid] [:= :id id]]]}
      sql/format))

(defn get-random [{gid :gid}]
  (-> (core/get-random {:columns [:kind :description :data]
                        :gid gid})))

(defn get-random-by-kind [{kind :kind gid :gid}]
  (-> (core/get-random-where {:columns [:kind :description :data]
                              :where [:= :kind kind]
                              :gid gid})
      sql/format))
