(ns marrano-bot.parse
  (:require [clojure.string :as s]))

(defn- get-command
  [text]
  (-> text
      (s/split #" ")
      first))


(defn command
  [text]
  (s/replace (get-command text) "!" ""))

(defn command?
  [text]
  (s/starts-with? (get-command text) "!"))

(defn parse
  [data]
  (let [matcher (re-matcher #"!\s*(?<cmd>[a-zA-Z0-9]+)\s*(?<text>.*)?" data)]
    (when (.matches matcher)
      (let [cmd       (s/lower-case (.group matcher "cmd"))
            predicate (.group matcher "text")]
        [(s/lower-case cmd) predicate]))))
