;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns moolite.bot.send
  (:require [config.core :refer [env]]
            [muuntaja.core :refer [encode]]
            [org.httpkit.client :as http]
            [redelay.core :refer [state]]
            [taoensso.timbre :as timbre :refer [debug]]))

(def token (:telegram-token env))
(def telegram-api (str "https://api.telegram.org/bot" token))
(def force-telegram-api (:telegram-api env))
(def webhook-url (str (:webhook env)
                      "/t/"
                      (:webhook-secret env)))

(defn- as-json
  [data]
  (->> data
       (encode "application/json")
       (slurp)))

(defn api [opts & fun]
  (let [{:keys [method]} opts
        payload (dissoc opts :method)]
    (http/post (str telegram-api "/"  method)
               {:headers {"Content-Type" "application/json"}
                :body (as-json payload)}
               (fn [resp]
                 (when-let [err (:error resp)] (timbre/error err))
                 (when fun (fun resp))))))

(def webhook (state :start
                    (when (:webhook-register env)
                      (api {:method "setWebhook"
                            :url webhook-url
                            :max_connections 100
                            :allowed_updates ["message" "callback_query"]}
                           (fn [{:keys [error]}]
                             (println "registered webhook" webhook-url))))
                    :stop
                    (when (:webhook-register env)
                      (api {:method "deleteWebhook"
                            :drop_pending_updates true}))))

(defn- as-message [data]
  (let [data (assoc data :parse_mode "MarkdownV2")]
    (when force-telegram-api
      (api data))
    data))

(defn photo [chat-id photo text]
  (debug ["photo" chat-id photo])
  (as-message {:method "sendPhoto"
               :chat_id chat-id
               :photo photo
               :caption text}))

(defn video [chat-id video caption]
  (debug ["video" chat-id video])
  (as-message {:method "sendVideo"
               :chat_id chat-id
               :video video
               :caption caption}))

(defn text [chat-id message_text]
  (debug ["text" chat-id message_text])
  (as-message {:method "sendMessage"
               :chat_id chat-id
               :text message_text}))

(defn location [chat-id lat lon]
  (debug ["location" chat-id lat lon])
  (as-message {:method "sendLocation"
               :chat_id chat-id
               :latitude lat
               :longitude lon}))

(defn dice [chat-id text]
  (debug ["dice" chat-id text])
  (as-message {:method "sendDice"
               :chat_id chat-id
               :text text}))

(defn links [chat-id text links]
  (debug ["links" chat-id links])
  (as-message {:method "sendMessage"
               :chat_id chat-id
               :text text
               :reply_markup (as-json {:inline_keyboard (map #(conj [] %) links)})}))
