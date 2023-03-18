;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns moolite.bot.send
  (:require [muuntaja.core :refer [encode]]
            [taoensso.timbre :as timbre :refer [debug]]))

(defn- as-json
  [data]
  (->> data
       (encode "application/json")
       slurp))

(defn- as-message [data]
  (merge {:disableNotification true
          :parse_mode "MarkdownV2"}
         data))

(defn photo [chat-id photo text]
  (debug "sending photo" '(chat-id photo))
  (as-message {:method "sendPhoto"
               :text text}))

(defn video [chat-id video caption]
  (debug "sending video" '(chat-id video))
  (as-message {:method "sendVideo"
               :video video
               :caption caption}))

(defn text [chat-id message_text]
  (debug "sending text" '(chat-id message_text))
  (as-message {:method "sendMessage"
               :text message_text}))

(defn location [chat-id lat lon]
  (debug "sending location" '(chat-id lat lon))
  (as-message {:method "sendLocation"
               :chat_id chat-id
               :latitude lat
               :longitude lon}))

(defn dice [chat-id text]
  (debug "sending dice" '(chat-id text))
  (as-message {:method "sendDice"
               :chat_id chat-id
               :text text}))

(defn links [chat-id text links]
  (debug "sending links" '(chat-id links))
  (as-message {:method "sendMessage"
               :chat_id chat-id
               :text text
               :reply_markup (as-json {:inline_keyboard (map #(conj [] %) links)})}))
