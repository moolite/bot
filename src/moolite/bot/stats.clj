;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns moolite.bot.stats
  (:require [clj-fuzzy.metrics :refer [dice]]
            [clojure.string :as string]))

(def ^:private stat-words
  [[:russacchiotta "russa" "russacchiotta"]
   [:polacchina "pupy" "pupa" "poska" "polacchina"]
   [:potta "figa" "potta" "signorina" "tette" "pheega"]
   [:aliemmo "aliemmo" "aliem" "lorenzo" "lallini" "lallo"]
   [:gatto "grumpycat" "gatto" "gattino" "bibbiano"]
   [:lukke "lukke" "luca" "bertolovski"]
   [:suppah "suppah" "supphppa" "suppahsrv" "munne"]
   [:amiga "amiga" "vampire" "cd32" "a1200" "a600" "acceleratore" "blitter" "terriblefire" "tf330" "tf530" "warp"]
   [:commodore "commodore" "c64" "vic20"]
   [:retro "spectrum" "speccy" "coleco" "atari" "falcon"]
   [:warez "warez" "crack" "key" "chiave"]
   [:bigdata "mongo" "elasticsearch" "elastic" "bigdata"]
   [:coin "bitcoin" "litecoin" "coin" "musk" "elon" "elon musk" "speculazione"]
   [:umme "umme" "ummme" "ummmeee" "umm3"]
   [:chiesa "dio" "papa" "chiesa" "religione" "porco"]
   [:mj "lsd" "mj" "marjuana" "erba" "pianta" "piantina"]
   [:marrano "marran" "marrans" "marrani" "mrrny"]
   [:brutt "grumpy" "grumpizza" "ummatore" "cattiv" "biden" "trump" "putin" "puteen" "moo" "pastoso" "lento" "manno" "mannosu" "brutto"]
   [:bell "brav" "bella" "ciccia" "pizza" "pasta" "vodka" "essi" "essisu" "massisu"]])

(defn calculate-distance
  [w1 w2]
  (dice w1 w2))

(defn calculate-rank-word
  [word words]
  (reduce
   #(max (calculate-distance word %2) %1)
   0 words))

(defn get-stats-from-phrase
  [word]
  (->> stat-words
       (filterv #(< 0.6 (calculate-rank-word word (rest %))))
       (map first)))

(defn get-all-stats
  [phrase]
  (let [words (string/split phrase #"\s+")]
    (->> words
         (mapcat get-stats-from-phrase))))
