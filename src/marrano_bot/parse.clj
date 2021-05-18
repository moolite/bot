;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns marrano-bot.parse
  (:require [clojure.string :as s]))

(def parse-regex
  #"[/a-z]*!?\s*(?<cmd>[a-zA-Z0-9]+)\s+(?<text>.+)?")

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
  (and (= \! (first data))
       (.matches (re-matcher parse-regex data))))
