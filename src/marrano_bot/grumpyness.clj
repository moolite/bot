(ns marrano-bot.grumpyness
  (:require [marrano-bot.stats :refer [get-all-stats]]
            [clojure.string :as string]))

(def score
  [[-3 :russacchiotta :polacchina :potta]
   [-2 :bell]
   [-1 :amiga :commodore :retro :warez]
   [1 :marrano :umme :mj]
   [2 :umme :brutt]
   [3 :bigdata :coin :chiesa]])

(defn calculate-thing-grumpyness
  [thing]
  (->>
   (map #(if (.contains % thing) (first %) 0)
        score)
   (reduce + 0)))

(defn calculate-grumpyness
  [phrase]
  (->> phrase
       (get-all-stats)
       (map calculate-thing-grumpyness)
       (reduce +)))
