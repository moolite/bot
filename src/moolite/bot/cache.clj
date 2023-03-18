(ns moolite.bot.cache)

;; A cache of the last x messager per-group
(def cache (atom []))

(defn add! [message]
  (if (> (count @cache) 500)
    (let [c (pop @cache)]
      (reset! cache c))
    (swap! cache conj message)))

(defn search-by [f]
  (filter @cache f))
