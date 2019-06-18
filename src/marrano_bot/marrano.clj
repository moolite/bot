(ns marrano-bot.marrano
  (:require [marrano-bot.parse :as p]
            [marrano-bot.db :as db]
            [morse.handlers :as h]
            [morse.api :as t]
            [config.core :refer [env]]
            [clojure.string :as s]))

;; Configuration
(def token
  (:token env))

(def webhook-url
  (str (:webhook env)
       "/t/"
       (:secret env)))

(defn init!
  []
  (do (println ">> registered webhook URL: " webhook-url)
      (t/set-webhook token webhook-url)
      (db/init!)))

(defn- template
  [tpl text]
  (s/replace tpl "%" text))

;; answer functions
(defn- rispondi
  [text]
  (let [[cmd pred] (p/parse text)
        tpl        (db/get-command cmd)]
    (if tpl (template tpl pred))))

;; Slap answers
(defn- slap
  [text]
  (let [text      (s/split text #" ")
        target    (get text 1)
        user-slap (s/join " " (drop 2 text))
        slap      (or (when (> 0 (.length user-slap))
                        user-slap)
                      (db/get-rand-slap))]
    (if (s/includes? slap "%")
      (str "@me " (template slap target))
      (str "@me slappa " target " con " slap))))

;; Slap save
(defn- slap-ricorda
  [text]
  (db/add-slap text))

;; Remember a new quote
(defn- ricorda
  [text]
  (let [[cmd pred] (p/parse text)]
    (if (and (not (nil? cmd))
             (not (empty? cmd)))
      (if (= cmd "slap")
        (slap-ricorda pred)
        (db/add-command cmd pred)))))

;; Remember one or more PhotoSize
(defn- ricorda-photo
  [id photos]
  (let [photo-ids (:file_id photos)]
    (db/update-at! [:photos] #(concat % photo-ids))))

;; Forget a quote
(defn- dimentica
  [text]
  (let [[cmd] (p/parse text)]
    (db/rem-command cmd [:custom cmd])))

;; Help message
(defn- paris-help
  []
  (let [list (->> (keys (db/get-in! [:commands]))
                  sort
                  (map #(str "- !" % "\n"))
                  (apply str))]
    (str "Helpy *paris*:\n\n" list)))

;; Request Handler
(h/defhandler bot-api
  (h/command "paris"
             {{id :id} :chat}
             (t/send-text token id {:parse_mode "Markdown"}
                          (paris-help)))

  (h/command "slap"
             {{id :id} :chat text :text}
             (t/send-text token id
                          (slap text)))

  (h/command "ricorda"
             {{id :id} :chat text :text}
             (do (ricorda text)
                 (t/send-text token id
                              "umme... ho imparato qualcosa!")))

  (h/command "dimentica"
             {{id :id} :chat text :text}
             (do (dimentica text)
                 (t/send-text token id
                              "non ricordo più")))

  (h/command "russacchiotta"
             {{id :id} :chat}
             (let [photo (rand-nth (db/get-in! [:photos]))]
               (t/send-photo token id photo)))

  ;; Commands message handler
  (h/message {{id :id chat-type :type} :chat text :text photo :photo}
             (cond
               (and text (p/command? text))
               (let [response (rispondi text)]
                 (when response
                   (t/send-text token id response)))

               (and photo (= chat-type "private"))
               (let [photo-id (ricorda-photo photo)]
                 (t/send-text token id (str "id: " photo-id)))))

  ;; Private photo messages
  (h/message {{id :id chat-type :type} :chat photo :photo}
             (when
                 (ricorda-photo id photo))))
