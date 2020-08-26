(ns marrano-bot.marrano
  (:require [marrano-bot.parse :as p]
            [marrano-bot.db :as db]
            [marrano-bot.search :refer [get-stats-from-phrase]]
            [morse.handlers :as h]
            [morse.api :as t]
            [muuntaja.core :as m]
            [config.core :refer [env]]
            [clojure.string :as s]
            [taoensso.timbre :as timbre :refer [info debug warn error]]))

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

;; Forget a quote
(defn- dimentica
  [text]
  (let [[cmd] (p/parse text)]
    (db/rem-command cmd [:custom cmd])))


;; Photo
(defn ricorda-photo
  [photo caption]
  (db/add-to-vec!
   [:photo]
   {:photo   photo
    :caption caption})
  "umme.")

(defn- dimentica-photo
  [photo]
  (db/del-from-vec!
   [:photo]
   #(= (:photo_id %) photo)))

;; Help message
(defn- paris-help
  []
  (let [list (->> (keys (db/get-in! [:commands]))
                  sort
                  (map #(str "- !" % "\n"))
                  (apply str))]
    (str "Helpy *paris*:\n\n" list)))

;; links
(defn- as-json
  [data]
  (->> data
       (muuntaja.core/encode "application/json")
       slurp))

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

;; Stats
;;
(defn update-stats
  [text]
  "increments stats based on spoken words"
  (loop [stats (seq (get-stats-from-phrase text))]
    (when (first stats)
      (debug "stats:" stats)
      (db/inc! :stats (first stats)))
    (when (next stats)
      (recur (next stats)))))

(defn get-stats
  []
  (->> (db/get-in! [:stats])
       (sort-by second)
       (map #(str "- `" (first %) "` :: *" (second %) "*"))
       (s/join "\n")))

;;
;; Responses
;;
(defn- send-message
  [parts]
  (merge {:method "sendMessage"
          :disableNotification true
          :parse_mode "Markdown"
          :text "qualcosa e' andato storto... colpa dei bot russi."}
         parts))

(defn- send-photo
  [parts]
  (merge {:method "sendPhoto"
          :disableNotification true}
         parts))

(defn bot-api
  [{{id :id chat-type :type} :chat caption :caption photo :photo text :text}]
  (when text
    (update-stats text))
  (let [caption (or caption "")
        text (or text "")]
    (cond (s/starts-with? text "/paris")
          (send-message {:chat_id id
                         :text (paris-help)})
          ;;
          ;; /slap
          ;;
          (s/starts-with? text "/slap")
          (send-message {:chat_id id
                         :text (slap text)})
          ;;
          ;; /ricorda
          ;;
          (s/starts-with? text "/ricorda")
          (do
            (ricorda text)
            (send-message {:chat_id id
                           :text "umme ... ho imparato *qualcosa*"}))
          ;;
          ;; /link | /nota
          ;;
          (some #(s/starts-with? text %)
                ["/link" "/nota"])
          (let [response (links text)]
            (when response
              (send-message {:chat_id id
                             :reply_markup (as-json {:inline_keyboard (map #(conj [] %) response)})
                             :text "ecco cosa ho trovato in *A-TEMP:*"})))

          ;;
          ;; Photos
          ;;
          (and photo caption
               (s/starts-with? caption "/ricorda"))
          (let [caption (or (last (re-find #"/ricorda (.*)"caption)) "")
                photo (:file_id (first photo))
                response (ricorda-photo photo caption)]
            (send-message {:chat_id id
                           :text (str "umme... " caption)}))

          (some #(s/starts-with? (s/lower-case text) %)
                ["/russa" "/pup" "pupy" "pupa" "russacchiotta" "signorina" "!russa" "!pupa" "potta"])
          (let [item (db/get-rand-in! [:photo])]
            (send-photo (merge {:chat_id id}
                               item)))

          ;;
          ;; Stats
          ;;
          (and text (s/starts-with? text "/stats"))
          (send-message {:chat_id id
                         :text (get-stats)})

          ;;
          ;; il resto
          ;;
          (and text (p/command? text))
          (let [response (rispondi text)]
            (when response
              (send-message {:chat_id id
                             :text response})))

          ;;
          ;; Stats
          ;;
          :else "")))
