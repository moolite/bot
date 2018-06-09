(ns marrano-bot.bot
  (:require [org.httpkit.client :as http]
            [cheshire.core :as j]
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
      (j/parse-string body))))

(defn- set-webhook-url
  [url]
  (http/post (str (:api env) "setWebhook")
             {:query-params {:url url}}
             (fn [{:keys [status headers body error]}] ;; asynchronous response handling
               (if error
                 (println "Failed, exception is " error)
                 (println "webhook registered: " status)))))

(defn send-message
  [body]
  (let [{:keys [status headers body error]} @(http/post (str (:api env) "sendMessage")
                                                        {:headers {"content-type" "application/json"}
                                                         :body (j/generate-string body)})]
    (if error
      (throw error)
      (j/parse-string body))))

(defn- parse-message
  [data]
  (let [data-text (s/split data #" ")
        cmd       (s/lower-case (first data-text))
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
        (s/replace text "%" (or user ""))))))

(defn answer-webhook
  "create a new message from the webhook"
  [data]
  (when (> (count data) 3)
    (let [message-id      (get-in data [:message :message_id])
          chat-id         (get-in data [:message :chat :id])
          [cmd predicate] (parse-message (get-in data [:message :text]))
          message         (answer cmd predicate)]
      (if message
        {:method               "sendMessage"
         :text                 message
         :chat_id              chat-id
         :reply_to_message_id  message-id
         :disable_notification true
         :parse_mode           "Markdown"}
        {}))))

(defn init
  "initialize the bot

   Registers a new webhook"
  []
  (let [{:keys [url]} (get-webhook-info)]
    (if (or (not url)
            (= url hook-url))
      (do (println "setting hook url as " hook-url)
          @(set-webhook-url hook-url)))))

(def asd
  {:update_id 82038896,
   :message {:message_id 2, :from {:id 318062977, :is_bot false, :first_name "lilo", :username "lilo060", :language_code "en-US"},
             :chat {:id 318062977, :first_name "lilo", :username "lilo060", :type "private"},
             :date 1528539408,
             :text "bot", :entities [{:offset 0, :length 6, :type "bot_command"}]}})

(answer-webhook asd)
