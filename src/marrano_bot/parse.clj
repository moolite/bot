(ns marrano-bot.parse
  (:require [clojure.string :as s]))

(def parse-regex
  #"!\s*(?<cmd>[a-zA-Z0-9]+)\s+(?<text>.+)?")

(defn parse
  [data]
  (let [matcher (re-matcher parse-regex data)]
    (when (.matches matcher)
      (let [cmd       (s/lower-case (.group matcher "cmd"))
            predicate (.group matcher "text")]
        [(s/lower-case cmd) predicate]))))

(defn- get-command
  [text]
  (-> text
      parse
      first))

(defn command
  [text]
  (s/replace (get-command text) "!" ""))

(defn command?
  [data]
  (let [matcher (re-matcher parse-regex data)]
    (.matches matcher)))
