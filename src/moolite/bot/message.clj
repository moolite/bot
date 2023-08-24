(ns moolite.bot.message
  (:require [clojure.string :as string]))

(def reserved-characters ["_" "*" "[" "]" "(" ")" "~" "`" ">" "#" "+" "-" "=" "|" "{" "}" "." "!"])
(def reserved-rex #"([_\*\[\]\(\)\~\`\>\#\+\-\=\|\{\}\.\!])")

(defn escape [s]
  (if s
    (string/replace s reserved-rex #(str "\\" (first %1)))
    ""))
