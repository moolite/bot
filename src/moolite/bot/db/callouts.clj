(ns moolite.bot.db.callouts
  (:require [moolite.bot.db.core :as core]
            [honey.sql :as sql]))

(def table :callouts)

(defn create-table []
  (-> {:create-table table
       :with-columns [[:callout [:varchar 128] [:not nil]]
                      [:text :text [:not nil]]
                      [:gid [:varchar 64] [:not nil]]
                      [[:primary-key :callout :gid]]
                      [[:foreign-key :gid]
                       [:references :groups]]]}
      (sql/format)))

(defn insert [{:keys [callout text gid]}]
  (-> {:insert-into table
       :columns [:gid :callout :text]
       :values [[gid callout text]]
       :on-conflict [:callout :gid]
       :do-update-set :text}
      (sql/format)))

(defn insert-many [data]
  (-> {:insert-into table
       :columns [:gid :callout :text]
       :values (map (fn [d] [(:gid d) (:callout d) (:text d)]) data)
       :on-conflict [:callout :gid]
       :do-update-set :text}
      (sql/format)))

(defn one-by-callout [{callout :callout gid :gid}]
  (-> {:select [:callout :text]
       :from table
       :where [:and
               [:= :callout callout]
               [:= :gid gid]]}
      (sql/format)))

(defn all-keywords [{gid :gid}]
  (-> {:select [:callout]
       :from table
       :where [:= :gid gid]}
      (sql/format)))

(defn delete-by-callout [callout]
  (-> {:detele-from table
       :where [:= :callout callout]}
      (sql/format)))

(defn get-random
  "get random command"
  [{:keys [gid]}]
  (-> {:table "callouts"
       :columns [:callout :text]
       :where [:= :gid gid]}
      (core/get-random)
      (sql/format)))
