;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns marrano-bot.marrano
  (:require [marrano-bot.parse :as p]
            [marrano-bot.db :as db]
            [marrano-bot.stats :refer [get-stats-from-phrase]]
            [marrano-bot.grumpyness :refer [calculate-grumpyness]]
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
;;


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
;;
(defn- paris-help
  []
  (let [list (->> (keys (db/get-in! [:commands]))
                  sort
                  (map #(str "- !" % "\n"))
                  (apply str))]
    (str "Helpy *paris*:\n\n" list)))

;; Link
;;
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
          [{:url (str "https://lmgtfy.com/?q=" text "&pp=1&s=d" "&s=l")
            :text "???? LMGIFY"}]
          results)))))

;; Grumpyness
;;
(defn update-grumpyness
  [username text]
  (when (and username text)
    (let [grumpyness (calculate-grumpyness text)]
      (db/inc-by! grumpyness :grumpyness username))))

(defn get-grumpyness
  []
  (->> (db/get-in! [:grumpyness])
       (sort-by second)
       (map #(str "- _" (first %) "_ *" (second %) "*"))
       (s/join "\n")))

(defn get-or-set-grumpyness
  [username text]
  (let [[_ name pts] (re-find #"\/[^\s]+ ([^\s]+) ([\+\-]?\d+)" text)
        pts (when (some? pts)
              (Integer/parseInt pts))
        name (when (and (db/get-in! [:grumpyness name])
                        (not= username name)) name)]
    (do (when (and name
                   pts)
          (db/inc-by! pts :grumpyness name))
        (get-grumpyness))))

;; Stats
;;
(defn update-stats
  [text]
  "increments stats based on spoken words"
  (doseq [stat (get-stats-from-phrase text)]
    (db/inc! :stats stat)))

(defn get-stats
  []
  (->> (db/get-in! [:stats])
       (sort-by second)
       (map #(str "- _" (name (first %)) "_ *" (second %) "*"))
       (s/join "\n")))

(defn prometheus-metrics
  []
  (->> (db/get-in! [:stats])
       (map (fn [[k v]] (str "# HELP marrano_" (name k) " metric\n"
                             "# TYPE marrano_" (name k) " gauge\n"
                             "marrano_" (name k) " " v)))
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
  [{{username :username} :from {id :id chat-type :type} :chat caption :caption photo :photo text :text}]
  (when (and text (not (s/starts-with? text "/")))
    (do (update-stats text)
        (update-grumpyness username text)))
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
          (let [caption (or (last (re-find #"/ricorda (.*)" caption)) "")
                photo (:file_id (first photo))
                response (ricorda-photo photo caption)]
            (send-message {:chat_id id
                           :text (str "umme... " caption)}))

          (some #(s/starts-with? (s/lower-case text) %)
                ["/russa" "/pup" "pupy" "pupa" "russa" "russia" "russacchiotta" "signorina" "!russa" "!pupa" "potta"])
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
          ;; Grumpyness
          ;;
          (and text (s/starts-with? text "/grumpy"))
          (send-message {:chat_id id
                         :text (get-or-set-grumpyness username text)})
          ;;
          ;; /ricorda
          ;;
          (s/starts-with? text "/ricorda")
          (do
            (ricorda text)
            (send-message {:chat_id id
                           :text "umme ... ho imparato *qualcosa*"}))
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
