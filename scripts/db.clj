(ns db
  (:require [clojure.edn :as edn]
            [redelay.core :refer [stop]]
            [moolite.bot.db :as db]
            [moolite.bot.db.abraxoides :as abraxoides]
            [moolite.bot.db.callouts :as callouts]
            [moolite.bot.db.groups :as groups]
            [moolite.bot.db.links :as links]
            [moolite.bot.db.media :as media]
            [moolite.bot.db.stats :as stats]))

(defn- spy [a]
  (println a)
  a)

(defn init-db [& _]
  (println "Initialising the bot.db sqlite database")
  (doseq [q [(groups/create-table)
             (abraxoides/create-table)
             (callouts/create-table)
             (links/create-table)
             (media/create-table)
             (stats/create-table)]]
    (println "Executing: " q)
    (let [res (db/execute-one! q)]
      (println res)))
  (stop))

(defn insert-group [{gid :gid}]
  (deref db/db)
  (-> (groups/insert {:gid gid :title "_"})
      (db/execute!)
      (println))
  (stop))

(defn photos-from-edn [{gid :gid}]
  (println "Importing *in* for group " gid)
  (deref db/db)
  (let [photos (-> (slurp *in*)
                   (edn/read-string)
                   :photo)]
    (doseq [p photos]
      (println "importing photo " p)
      (-> (media/insert {:data [(:photo p)]
                         :kind "photo"
                         :text (:caption p)
                         :gid gid})
          (db/execute!)
          (println))))
  (stop))

(defn commands-from-edn [{gid :gid}]
  (deref db/db)
  (let [commands (-> (slurp *in*)
                     (edn/read-string)
                     :commands)
        data (map (fn [c] {:gid gid
                           :callout (first c)
                           :text (second c)})
                  commands)]
    (-> (callouts/insert-many data)
        (db/execute!)
        (println)))
  (stop))
