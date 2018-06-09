(ns marrano-bot.bot
  (:require [org.httpkit.client :as http]
            [jsonista.core :as j]
            [clojure.string :as s]
            [config.core :refer [env]]
            [codax.core :as c]))

;; Configuration
(def api-url
  (:api env))

(def hook-url
  (str (get-in env [:hook :url])
       (get-in env [:hook :token])))

;; persistence layer
(def db
  (c/open-database! "data"))

;;; initial seed data
(let [slaps (c/get-at! db [:slaps])]
  (if (empty? slaps)
    (c/assoc-at! db [:slaps]
                 {"marrano" "%, sei un marrano!"
                  "schif"   "%, io ti schifo!"
                  "strunz"  "%, sei strunz!"
                  "paris"   "%, sei più helpy di paris hilton!"
                  "bot"     "mannò, massù, sù!"})))

;; REST API
(defn- get-webhook-info
  []
  (let [{:keys [status headers body error]} @(http/get (str api-url "getWebhookInfo"))]
    (if error
      (throw error)
      (j/read-value body))))

(defn- set-webhook-url
  [url]
  (http/post (str (:api env) "setWebhook")
             {:query-params {:url url}}
             (fn [{:keys [status headers body error]}] ;; asynchronous response handling
               (if error
                 (println "Failed, exception is " error)
                 (println "webhook registered: " status)))))

(defn- parse-message
  [data]
  (let [data-text (s/split (:text data) #" ")
        cmd       (first data-text)
        predicate (s/join " " (rest data-text))]
    [cmd predicate]))

(defn- ricorda
  "Add custom message"
  [data]
  (let [[cmd predicate] (parse-message data)]
    (c/assoc-at! db [:custom cmd] predicate)
    nil))

(defn- dimentica
  "Remove custom message"
  [data]
  (let [[cmd predicate] (parse-message data)]
    (c/dissoc-at! db [:custom cmd])
    nil))

(defn- answer
  [cmd user]
  (condp = cmd
    "ricorda"   (ricorda user)
    "dimentica" (dimentica user)

    (let [text (or (c/get-at! db [:slaps cmd])
                   (c/get-at! db [:custom cmd]))]
      (if text
        (s/replace text "%" user)))))

(defn answer-webhook
  "create a new message from the webhook"
  [data]
  (let [message-id      (:message_id data)
        chat-id         (:chat_id data)
        [cmd predicate] (parse-message data)
        message         (answer cmd predicate)]
    (if message
      {:method              "sendMessage"
       :chat-id             (:chat_id    data)
       :reply-to-message-id (:message_id data)
       :text                message}
      {})))

(defn init
  "initialize the bot

   Registers a new webhook"
  []
  (let [{:keys [url]} (get-webhook-info)]
    (if (or (not url)
            (= url hook-url))
      (do (println "setting hook url as " hook-url)
          @(set-webhook-url hook-url)))))
