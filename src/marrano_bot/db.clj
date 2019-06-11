(ns marrano-bot.db
  (:require [clojure.edn :as edn]
            [clojure.java.io :as io]))

(def db-filename "./db.edn")
(def db
  (atom {:commands {"marrano"   "%, sei un marrano!",
                    "schif"     "%, io ti schifo!",
                    "betaschif" "%, io ti betaschifo!",
                    "strunz"    "%, sei strunz!",
                    "paris"     "%, sei più helpy di paris hilton!",
                    "chain"     "%, sei più lento di una blockchain!",
                    "cripto"    "%, ti criptobottokremlino!",
                    "soviet"    "%, ti mando a Stalingrado!",
                    "russa"     "%, deh or dico a Putin di tolgliert le russacchiotte di man!",
                    "spec"      "%, ti fo crashare pur di non cambiare la mia spec.",
                    "bot"       "mannò, massù, sù!"}
         :slap ("una grande trota!"
                "le diciotto bobine edizione limitata de La Corazzata Potemkin durante Italia Inghilterra")}))

(defn save! []
  (spit db-filename (prn-str @db)))

(defn load! []
  (reset! @db (edn/read-string (slurp db-filename))))

(defn init!
  []
  (if (.exists (io/as-file db-filename))
    ; load the db from file
    (load!)
    ; save the default db to file
    (save!))
  ; add the atom watcher
  (add-watch db :save save!))


(defn get-in
  [path]
  (get-in @db path))

(defn update-in
  [k v]
  (swap! db update-in k v))
