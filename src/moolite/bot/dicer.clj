;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns moolite.bot.dicer
  (:require [instaparse.core :as insta]
            [clojure.edn :as edn]
            [clojure.string :as string]))

(def die-lang
  (insta/parser
   "die    : throws <'d'> dice (<'k'> keep)?
    throws : #'[0-9]+'
    dice   : #'[0-9]+'
    keep   : #'[0-9]+'
    "))

(defn parse-die [str]
  (insta/transform
   {:throws edn/read-string
    :dice edn/read-string
    :keep edn/read-string}
   (die-lang str)))

(defn roll
  [raw]
  (let [[_ throws dice keep] (parse-die raw)
        results (map (fn [_] (+ (inc (rand-int  dice))))
                     (range throws))]
    (if (some? keep)
      (->> (sort > results)
           (take keep)
           vec)
      [(reduce + results)])))

(defn as-emoji [lst]
  (->> lst
       (map str)
       (map #(string/replace % #"(\d)" "$1\u20E3"))))

(comment
  (roll "2d20") ;; => 12
  (roll "4d6k2") ;; => (6 4)
  (roll "1d20") ;; => 6
  (roll "8d10k6")) ;; => (10 9 7 4 3 2)
