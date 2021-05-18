;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns marrano-bot.stats
  (:require [clj-fuzzy.metrics :refer [dice]]
            [clojure.string :as string]))

(def stat-words
  [[:russacchiotta "russa" "russacchiotta"]
   [:polacchina "pupy" "pupa" "poska" "polacchina"]
   [:lukke "lukke" "luca" "bertolovski"]
   [:suppah "suppah" "supphppa" "suppahsrv" "munne"]
   [:gatto "grumpycat" "gatto" "gattino" "bibbiano"]
   [:aliemmo "aliemmo" "aliem" "lorenzo" "lallini"]
   [:marrano "marran" "marrans" "marrani" "mrrny"]
   [:amiga "amiga" "vampire" "cd32" "a1200" "a600" "acceleratore" "blitter" "aga" "terriblefire" "tf330" "tf530" "warp"]
   [:commodore "commodore" "c64" "vic20"]
   [:retro "spectrum" "speccy" "coleco" "atari" "falcon"]
   [:umme "umme" "ummme" "ummmeee" "umm3"]
   [:potta "figa" "potta" "signorina" "tette"]
   [:bigdata "mongo" "elasticsearch" "elastic" "bigdata"]])

(defn calculate-distance
  [w1 w2]
  (dice w1 w2))

(defn calculate-rank-word
  [word words]
  (reduce
   #(max (calculate-distance word %2) %1)
   0 words))

(defn calculate-rank
  [phrase words]
  (->> (string/split phrase #"\s+")
       (map #(calculate-rank-word % words))
       (apply max)))

(defn get-stats-from-phrase
  [phrase]
  (->> stat-words
       (filterv #(< 0.4 (calculate-rank phrase
                                        (rest %))))
       (map first)
       (reduce conj #{})))
