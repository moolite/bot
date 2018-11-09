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


;; seed database
(c/assoc-at! db [:commands]
             {"marrano"  "%, sei un marrano!",}
             "schif"     "%, io ti schifo!",
             "betaschif" "%, io ti betaschifo!",
             "strunz"    "%, sei strunz!",
             "paris"     "%, sei più helpy di paris hilton!",
             "chain"     "%, sei più lento di una blockchain!",
             "cripto"    "%, ti criptobottokremlino!",
             "soviet"    "%, ti mando a Stalingrado!",
             "russa"     "%, deh or dico a Putin di tolgliert le russacchiotte di man!",
             "spec"      "%, ti fo crashare pur di non cambiare la mia spec.",
             "bot"       "mannò, massù, sù!")

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

(defn- ricorda
  [text]
  (let [textlist (rest (s/split #" " text))
        cmd (first textlist)
        pred (s/join " " (rest textlist))]
    (c/update-at! db [:custom cmd] pred)))

(defn- dimentica
  [text]
  (let [cmd (first (rest (s/split #" " text)))]
    (c/dissoc-at! db [:custom cmd])))

(defn- slap
  [text]
  (let [slap (rand-nth (c/seek-at! db [:slap]))]
    (str "@me slappa " text " con " (second slap))))

(defn- slap-ricorda
  [text]
  (c/assoc-at! db [:slap] text))

;; Request Handler
(h/defhandler bot-api
  (h/command "paris"
             {{id :id} :chat}
             (let [commands (c/seek-at! db [:commands])
                   custom   (c/seek-at! db [:custom])
                   command-list (apply str (map #(str "- !" (first %) "\n"))
                                       (concat commands custom))]
               (t/send-text token id
                            {:parse_mode "Markdown"}
                            (str "Helpy *paris*:\n\n" command-list))))

  (h/command "ricorda"
             {{id :id} :chat text :text}
             (do (ricorda text)
                 (t/send-text token id "umme... ho imparato qualcosa!")))

  (h/command "dimentica"
             {{id :id} :chat text :text}
             (do (dimentica text)
                 (t/send-text token id "non ricordo più")))

  (h/command "slap"
             {{id :id} :chat text :text}
             (t/send-text token id (slap text)))

  (h/command "ricorda-slap"
             {{id :id} :chat text :text}
             (do (slap-ricorda text)
                 (t/send-text token id (str "interessante... " text))))

  (h/message {{id :id} :chat text :text}
             (when (and text (command? text))
               (let [response (rispondi text)]
                 (when response
                   (t/send-text token id (rispondi text)))))))

;; (bot-api {:message {:text "!paris hilton" :chat {:id 1234567}}})
