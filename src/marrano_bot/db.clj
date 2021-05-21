;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns marrano-bot.db
  (:require [clojure.edn :as edn]
            [clojure.java.io :as io]
            [clojure.string :as s]))

(def db-filename "./db.edn")
(def db
  (atom {:commands {"acbs"      "%, ti acbsizzo!"
                    "audi"      "%, ti installo i sensori posteriori, schermo, finestrini elettronici con massaggio e autoradio con bitcoin miner di una Audi nella Giulia!"
                    "azimuth"   "%, registra quella maledetta testinaaaaaaaa!!!!!!!!"
                    "betaschif" "%, io ti betaschifo!"
                    "bell"      "%, sei !bell"
                    "bloc"      "%, sei alpha.... Anzi pre-dev-unstable-preview-poc!"
                    "bot"       "mannò, massù, sù!"
                    "brav"      "%, sei marran, ma anche !brav"
                    "chain"     "%, sei più lento di una blockchain!"
                    "chrome"    "oh no! % e' diventato un google-coin miner!!!"
                    "ciocco"    "%, ti fo cioccar come il gatto!"
                    "cripto"    "%, ti criptobottokremlino!"
                    "filtro"    ".im_message_photo_thumb{filter:blur(15px)} .im_message_photo_thumb:hover{filter:blur(0)}"
                    "galoppo"   "%, corri ragazzo laggiù corri al galoppo orsuuuuu"
                    "gelato"    "%, meglio un culo gelato che un gelato nel culo"
                    "gmt"       "oh no! %, hai il fuso orario del gatto!"
                    "hz"        "%, ti sei incastrato in una scanline!?!?!!!oneoneoe"
                    "kb"        "%, devi finire la tastierina slim per l'ipaddo!"
                    "lallini"   "%, sei un massacratore di lallini!"
                    "lib"       "%, sei un ordo-lib-tard-marran!"
                    "lubrano"   "%, massù! Smetti di toccarti davanti alle repliche di 'mi manda Lubrano'!"
                    "marrana"   "%, sei un marrano femmina!"
                    "marrani"   "%, siete dei marrani!"
                    "marrano"   "%, sei un marrano!"
                    "mostro"    "%, sei diventato il mostro di Livorno!"
                    "nas"       "%, ti mando i NAS a controllarti il NAS!!"
                    "nvme"      "%, non hai 8 slot NVME pro a 20Tbps 1nm ultraa?????!!!!"
                    "okr"       "%, serve definire i nuovi BO del OKR con OBMS e OGM per il KPI nel BI per aumentare l'Intelligence della Dashboard"
                    "paris"     "%, sei più helpy di paris hilton!"
                    "piaga"     "%, sei peggio di uno ShInKuRo attaccato ai coglioni!!!oneone11"
                    "power"     "%, ti ricarico con la batteria di una Alfa Giulia 3400cc benzina!"
                    "retro"     "%, sei un marrano retrò!"
                    "rosso"     "%, sei un rossobruno!!11oneone"
                    "russa"     "%, deh or dico a Putin di tolgliert le russacchiotte di man!"
                    "schif"     "%, io ti schifo!"
                    "schioppo"  "%, te lo inculo con schioppo al galoppo!"
                    "seghe"     "%, troppe seghe ti dimentichi anche come ti chiamiii!!!oneoneone"
                    "silenti"   "Marrani! Siete più silenziosi del silenzio degli innocenti!"
                    "soviet"    "%, ti mando a Stalingrado!"
                    "sovrano"   "%, sei un sovraninsto-stalino-comunisto-criptorossobruno!"
                    "spec"      "%, ti fo crashare pur di non cambiare la mia spec."
                    "strunz"    "%, sei strunz!"
                    "tentacolo" "%, ti lascio solo un attimo e mi fai come i tentacoli di day of tentacle!"
                    "trap"      "%, hai ascoltato 8 ore di trap ed assistito a tutti i concerti di sferaebbasta?!?"
                    "vegano"    "%, sei diventato un LGBTVEGan!"
                    "vw"        "%, mi vuoi gasare con la vw???? Spegnilaaaa"}
         :slap ["una grande trota!"
                "le diciotto bobine edizione limitata de La Corazzata Potemkin durante Italia Inghilterra"]
         :photos []
         :links []}))

(defn save!
  ([_ _ _ _]
   (save!))
  ([]
   (spit db-filename (prn-str @db))))

(defn load! []
  (reset! db (edn/read-string (slurp db-filename))))

(defn init!
  []
  (if (.exists (io/as-file db-filename))
    ; load the db from file
    (load!)
    ; save the default db to file
    (save!))
  ; add the atom watcher
  (add-watch db :save save!))

(defn get-in!
  [path]
  (get-in @db path))
(defn get-rand-in!
  [path]
  (rand-nth (get-in! path)))

(defn update-at!
  [k f]
  "update a path of an atom
   k key
   f update function"
  (swap! db (fn [val] (update-in val k f))))

(defn apply-in!
  [k fun default-value]
  (swap! db (fn [val] (update-in val k
                                 #(if % (fun %) default-value)))))

(defn add-to-vec!
  [k thing]
  "append a value to a db vector k
   k     key
   thing appended value"
  (update-at! k #(->> (conj % thing)
                      (filter some?)
                      vec
                      distinct)))

(defn del-from-vec!
  [k fun]
  "remove a thing from a vector using a filter function"
  (update-at! k (fn [val] (filterv #(not (fun %)) val))))

;;
;; High-level helpers
;;

;; Slaps
(defn add-slap [text]
  (add-to-vec! [:slap] text))
(defn get-slaps []
  (:slap @db))
(defn get-rand-slap []
  (get-rand-in! [:slap]))

;; commands
(defn add-command [name text]
  (swap! db (fn [val] (update-in val [:commands name] (fn [_] text)))))
(defn get-command [k]
  (get-in @db [:commands k]))
(defn rem-command [k]
  (update-at! [:commands]
              #(dissoc % k)))

;; Links
(defn add-link [url text]
  (add-to-vec! [:links]
               {:url   url
                :text text}))

(defn search-link [terms]
  (->> (get-in! [:links])
       (filterv #(s/includes? terms (:text %)))))

(defn rem-link [url]
  (update-at! [:links]
              (fn [links] (filterv #(not (= (:url %) url))
                                   links))))

(defn inc!
  [& k]
  (apply-in! k inc 1))

(defn inc-by!
  [n & k]
  (apply-in! k #(+ % n) 1))
