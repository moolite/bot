(ns marrano-bot.db
  (:require [clojure.edn :as edn]
            [clojure.java.io :as io]))

(def db-filename "./db.edn")
(def db
  (atom {:commands {"marrano"   "%, sei un marrano!"
                    "brav"      "%, sei marran, ma anche !brav"
                    "schif"     "%, io ti schifo!"
                    "betaschif" "%, io ti betaschifo!"
                    "strunz"    "%, sei strunz!"
                    "paris"     "%, sei più helpy di paris hilton!"
                    "chain"     "%, sei più lento di una blockchain!"
                    "cripto"    "%, ti criptobottokremlino!"
                    "soviet"    "%, ti mando a Stalingrado!"
                    "russa"     "%, deh or dico a Putin di tolgliert le russacchiotte di man!"
                    "spec"      "%, ti fo crashare pur di non cambiare la mia spec."
                    "acbs"      "%, ti acbsizzo!"
                    "bot"       "mannò, massù, sù!"
                    "silenti"   "Marrani! Siete più silenziosi del silenzio degli innocenti!"
                    "vegano"    "%, sei diventato un LGBTVEGan!"
                    "tentacolo" "%, ti lascio solo un attimo e mi fai come i tentacoli di day of tentacle!"
                    "seghe"     "%, troppe seghe ti dimentichi anche come ti chiamiii!!!oneoneone"
                    "rosso"     "%, sei un rossobruno!!11oneone"
                    "sovrano"   "%, sei un sovraninsto-stalino-comunisto-criptorossobruno!"
                    "piaga"     "%, sei peggio di uno ShInKuRo attaccato ai coglioni!!!oneone11"
                    "mostro"    "%, sei diventato il mostro di Livorno!"
                    "hz"        "%, ti sei incastrato in una scanline!?!?!!!oneoneoe"
                    "kb"        "%, devi finire la tastierina slim per l’ipaddo!"
                    "gmt"       "oh no! %, hai il fuso orario del gatto!"
                    "ciocco"    "%, ti fo cioccar come il gatto!"
                    "lallini"   "%, sei un massacratore di lallini!"
                    "azimuth"   "%, registra quella maledetta testinaaaaaaaa!!!!!!!!"
                    "filtro"    ".im_message_photo_thumb{filter:blur(15px)} .im_message_photo_thumb:hover{filter:blur(0)}"
                    "gelato"    "%, meglio un culo gelato che un gelato nel culo"}
         :slap ["una grande trota!"
                "le diciotto bobine edizione limitata de La Corazzata Potemkin durante Italia Inghilterra"]}))

(defn save! []
  (spit db-filename (prn-str @db)))

(defn load! []
  (reset! @db (edn/read-string (slurp db-filename))))

(defn init!
  []
  (if (.exists (io/as-file db-filename))
    ; load the db from file
    (load!)
    ; save the default db to file
    (save!))
  ; add the atom watcher
  (add-watch db :save save!))


(defn get-in
  [path]
  (get-in @db path))

(defn update-in
  [k f]
  (swap! db update-in k f))
