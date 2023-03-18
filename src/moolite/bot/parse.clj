;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns moolite.bot.parse
  (:require [instaparse.core :as insta]
            [taoensso.timbre :as timbre :refer [info debug]]))

(def text-lang
  (insta/parser
   "
    message   : space* (command|callout)? abraxas (space rest)?
    <rest>    : subcmd space+ url | subcmd space+ url space+ text | url space+ text | url | text
    <subcmd>  : del|add
    <space>   : <' '>
    command   : <'/'>
    callout   : <'!'>
    abraxas   : #'[-a-zA-Z0-9]+'
    del       : <'rm'|'rem'|'del'|'rimuovi'|'cancella'|'dd'|'-'>
    add       : <'add'|'aggiungi'|'nuovo'|'crea'|'new'|'+'>
    url       : #'https?://[^ ]+'
    text      : !<'http'> #'.+'
    "))

(defn parse-message
  [{:keys [text]}]
  (let [results (insta/parse text-lang text :optimize :memory)]
    (if (insta/failure? results)
      []
      results)))
