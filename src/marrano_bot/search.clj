(ns marrano-bot.search
  (:require [clj-fuzzy.metrics :refer [dice]]))

(dice "bot" "un bot")
(dice "bot" "un bot e un marrano")

(def words
  ["russacchiotte" "libri" "marrani" "marrano" "asimov" "filosofia" "amiga" "AmigaOS"])

(def frasi
  [{:text "russacchiotte fighe"}
   {:text "libri"}
   {:text "marrani"}
   {:text "polacchine"}
   {:text "figa"}
   {:text "libri gratuiti"}
   {:text "sconti amazon"}
   {:text "acceleratori amiga"}
   {:text "AOS 3.1.4"}
   {:text "AmigaOS 3.9"}])

(take 5 (filterv #(< 0 (dice "amiga" %))
                 words))
 
(defn take-topmost
  [term]
  (->> words
       (filterv #(< 0 (dice term %)))
       (take 5)
       (vec)))

(defn calculate-rank
  [word]
  (->> (map (fn [x] {:text (:text x) :rank (dice word (:text x)) :term word})
            frasi)
       (filterv #(< 0.2 (:rank %)))))

(calculate-rank "amiga")

(dice "acceleratori amiga" "amiga")

(defn search
  []
  (->> (map calculate-rank words)))

(search)

;; - (each word)
;; - rank by word
;; - filtra per (> 0 x)
;; - 
