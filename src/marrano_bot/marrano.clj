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
  [id photo-sizes]
  (let [photo-id (:file_id (last photo-sizes))]
    (db/add-to-vec [:photos] photo-id)
    {:markdown true
     :text (str "ricorderò l'id :`" photo-id "`")}))

(defn- ricorda-video
  [id video])

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

;; links
(defn links
  [text]
  (let [[_ cmd tags] (re-find #"/[\w]+\s+([^\s]+)?\s+(.+)?" text)]
    (cond
      (and (nil? cmd)
           (nil? tags))
      (vec (db/get-in! [:links]))
      
      (and (some? cmd)
           (some? tags)
           (s/starts-with? cmd "http"))
      (do (db/add-link cmd tags)
          [{:url cmd
            :text tags}])

      (or (= cmd "rm")
          (= cmd "del")
          (= cmd "rimuovi")
          (= cmd "schifa"))
      (do (db/rem-link tags)
          [{:url tags
            :text "eliminato!"}])

      :else
      (let [results (db/search-link (rest (s/split text #"\s+")))]
        (if (empty? results)
          [{:url (str "https://lmgtfy.com/?q=" text "&pp=1&s=d""&s=l")
            :text "🖕 LMGIFY"}]
          results)))))

(defn- send-message
  [parts]
  (merge {:method "sendMessage"
          :disableNotification true
          :parse_mode "Markdown"
          :text "qualcosa e' andato storto... colpa dei bot russi."}
         parts))

(defn bot-api
  [{{id :id chat-type :type} :chat text :text}]
  (cond (s/starts-with? text "/paris")
        (send-message {:chat_id id :text (paris-help)})
        ;;
        ;; /slap
        ;; 
        (s/starts-with? text "/slap")
        (send-message {:chat_id id :text (slap text)})
        ;;
        ;; /ricorda
        ;; 
        (s/starts-with? text "/ricorda")
        (do
          (ricorda text)
          (send-message {:chat_id id :text "umme ... ho imparato *qualcosa*"}))
        ;;
        ;; /link | /nota
        ;;
        (or (s/starts-with? text "/link")
            (s/starts-with? text "/nota"))
        (let [response (links text)]
          (when response
            (send-message {:chat_id id
                           :reply_markup response
                           :text "ecco cosa ho trovato in *A-TEMP:*"})))
        ;;
        ;; il resto
        ;;
        (and text (p/command? text))
        (let [response (rispondi text)]
            (when response
              (send-message {:text response})))
        :else ""))

(comment
  (bot-api {:text "/link"})

  (links "/link https://example.com ex eg")
  (links "/link del https://example.com")
  (links "/link ex")
  (links "/link")

  (re-find #"/[\w]+ ([^\s]+) (.+)"
          "/ass foo bar"))

