(ns marrano-bot.marrano
  (:require [morse.handlers :as h]
            [morse.api :as t]
            [config.core :refer [env]]
            [codax.core :as c]
            [clojure.string :as s]))

;; Configuration
(def token
  (:token env))

(def webhook-url
  (str (:webhook env)
       "/t/"
       (:secret env)))

;; Database
(def db
  (c/open-database! (or (:db env)
                        "./db")))


(defn init!
  []
  (do (println ">> registered webhook URL: " webhook-url)
      (t/set-webhook token webhook-url)
      ;; seed database
      (c/assoc-at! db [:commands]
                   {"marrano"   "%, sei un marrano!",
                    "schif"     "%, io ti schifo!",
                    "betaschif" "%, io ti betaschifo!",
                    "strunz"    "%, sei strunz!",
                    "paris"     "%, sei più helpy di paris hilton!",
                    "chain"     "%, sei più lento di una blockchain!",
                    "cripto"    "%, ti criptobottokremlino!",
                    "soviet"    "%, ti mando a Stalingrado!",
                    "russa"     "%, deh or dico a Putin di tolgliert le russacchiotte di man!",
                    "spec"      "%, ti fo crashare pur di non cambiare la mia spec.",
                    "bot"       "mannò, massù, sù!"})))

;; answer functions
(defn- command?
  [text]
  (-> text
      (s/split #" ")
      first
      (s/starts-with? "!")))

(defn- parse-text
  [data]
  (let [matcher (re-matcher #"!\s*(?<cmd>[a-zA-Z]+)\s*(?<text>.*)?" data)]
    (when (.matches matcher)
      (let [cmd       (s/lower-case (.group matcher "cmd"))
            predicate (.group matcher "text")]
        [(s/lower-case cmd) predicate]))))

(defn- rispondi
  [text]
  (let [[cmd pred] (parse-text text)
        tpl        (or (c/get-at! db [:custom cmd])
                       (c/get-at! db [:commands cmd]))]
    (if tpl (s/replace tpl "%" pred))))
(rispondi "!marrano suppah")

(defn- slap
  [text]
  (let [slap   (rand-nth (c/seek-at! db [:slap]))
        target (apply str (rest (s/split text #" ")))]
    (str "@me slappa " target " con " (second slap))))

(defn- slap-ricorda
  [text]
  (c/assoc-at! db [:slap] text))

(defn- ricorda
  [text]
  (let [textlist (rest (s/split text #" "))
        cmd      (first textlist)
        pred     (s/join " " (rest textlist))]
    (if (= cmd "slap")
      (slap-ricorda pred)
      (c/assoc-at! db [:custom cmd] pred))))

(defn- dimentica
  [text]
  (let [cmd (first (rest (s/split text #" ")))]
    (c/dissoc-at! db [:custom cmd])))

(defn- paris-help
  []
  (let [commands (c/seek-at! db [:commands])
        custom   (c/seek-at! db [:custom])
        list     (map #(str "- !" (first %) "\n")
                      (concat commands custom))]
    (str "Helpy *paris*:\n\n"
         (apply str list))))

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

  (h/message {{id :id} :chat text :text}
             (when (and text (command? text))
               (let [response (rispondi text)]
                 (when response
                   (t/send-text token id response))))))

;; (bot-api {:message{:chat{:id 123} :text "/paris"}})
