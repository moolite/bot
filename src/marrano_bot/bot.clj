(ns marrano-bot.bot
  (:require [org.httpkit.client :as http]
            [clojure.string :as s]))

;; REST API

(def BASE_DOMAIN
  (System/getenv "BOT_DOMAIN"))

(def BASE_URL
  (let [token (System/getenv "TOKEN")]
    (str "https://api.telegram.org/bot" token "/")))

(defn register-webhook
  [token url]
  (http/post (str BASE_URL "setWebhook") 
             { :query-params {:url (str BASE_DOMAIN "/t/" token)} }))

(defn answer
  [cmd predicate]
  (condp = cmd
    "marrano" (str predicate ", sei un marrano!")
    "schif"   (str predicate ", io ti schifo!")
    "strunz"  (str predicate ", sei strunz!")
    "paris"   (str predicate ", sei più helpy di paris hilton!")
    "bot"     "mannò, massù, sù!"
    false))

(defn answer-webhook
  [data]
  (let [message-id (:message_id data)
        chat-id    (:chat_id data)
        data-text  (s/split (:text data) #" ")
        cmd        (first data-text)
        predicate  (apply str (rest data-text))
        message    (answer cmd predicate)]
    (if message
      {:method              "sendMessage"
       :chat-id             (:chat_id    data)
       :reply-to-message-id (:message_id data)
       :text                message}
      {})))
