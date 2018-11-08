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
  (c/open-database! (or (:db env
                             "./db"))))


;; seed database
(c/assoc-at! db [:commands]
  {"marrano" "%, sei un marrano!",
   "schif" "%, io ti schifo!",
   "betaschif" "%, io ti betaschifo!",
   "strunz" "%, sei strunz!",
   "paris" "%, sei più helpy di paris hilton!",
   "chain" "%, sei più lento di una blockchain!",
   "cripto" "%, ti criptobottokremlino!",
   "soviet" "%, ti mando a Stalingrado!",
   "russa" "%, deh or dico a Putin di tolgliert le russacchiotte di man!",
   "spec" "%, ti fo crashare pur di non cambiare la mia spec.",
   "bot" "mannò, massù, sù!"})

;; answer functions
(defn- command?
  [text]
  (-> text
      (s/split #" ")
      first
      (s/starts-with? "!")))

(defn- parse-text
  [data]
  (let [matcher (re-matcher #"!\s*(?<cmd>[a-zA-Z]+) (?<text>.*)" data)]
    (if (.matches matcher)
      (let [cmd (s/lower-case (.group matcher "cmd"))
            predicate (.group matcher "text")]
        [(s/lower-case cmd) predicate])
      [nil nil])))

(defn- rispondi
  [text]
  (let [[cmd pred] (parse-text text)
        tpl (or (c/get-at! db [:commands cmd]) (c/get-at! db [:custom cmd]))]
    (if tpl (s/replace tpl "%" pred))))

(defn- ricorda
  [text]
  (let [textlist (rest (s/split #" " text))
        cmd (first textlist)
        pred (rest textlist)]
    (c/update-at! db [:custom cmd] pred)))

(defn- dimentica
  [text]
  (let [cmd (first (rest (s/split #" " text)))]
    (c/dissoc-at! db [:custom cmd])))

;; Request Handler
(h/defhandler bot-api
  (h/command "paris"
             {{id :id} :chat}
             (let [command-list (apply str (map #(str "- !" (first %) "\n")
                                                (c/seek-at! db [:commands])))]
               (t/send-text token
                            id
                            {:parse_mode "Markdown"}
                            (str "Helpy paris:\n\n" command-list))))
  (h/command "ricorda"
             {{id :id} :chat text :text}
             (do (ricorda text)
                 (t/send-text token id "umme... ok")))
  (h/command "dimentica"
             {{id :id} :chat text :text}
             (do (dimentica text)
                 (t/send-text token id "dimenticai...")))
  (h/message {{id :id} :chat text :text}
             (if (and text
                      (command? text))
               (t/send-text token id (rispondi text)))))

;; (bot-api {:message {:text "!paris hilton" :chat {:id 1234567}}})
