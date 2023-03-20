;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns moolite.bot.actions
  (:require [clojure.string :as string]
            [clojure.core.match :refer [match]]
            [taoensso.timbre :as timbre :refer [spy info debug]]
            [moolite.bot.send :as send]
            [moolite.bot.dicer :as dicer]
            [moolite.bot.db :as db]
            [moolite.bot.db.groups :as groups]
            [moolite.bot.db.callouts :as callouts]
            [moolite.bot.db.stats :as stats]
            [moolite.bot.db.links :as links]
            [moolite.bot.db.media :as media]
            [moolite.bot.db.abraxoides :as abraxoides]
            [moolite.bot.message :as message]))

(defn register-channel [{{gid :id title :title} :chat}]
  (debug {:fn "register-channel" :gid gid :title title})
  (let [chan (-> (groups/insert {:gid gid :title title})
                 (db/execute-one!))]
    (send/text gid (printf "channel registered with id %d and title '%s'" (:gid chan) (:title chan)))))

(defn help [gid]
  (debug {:fn "help" :gid gid})
  (let [callout (->> (callouts/all-keywords {:gid gid})
                     (db/execute!)
                     (map :callout)
                     (map #(str "- " %))
                     (string/join "\n"))]
    (send/text gid callout)))

(defn grumpyness [gid]
  (debug ["grumpyness" gid])
  (let [grumpyness (-> (stats/all {:gid gid})
                       (db/execute-one!))]
    (send/text gid grumpyness)))

(defn link-add [gid url & text]
  (debug ["link-add" gid])
  (-> (links/insert {:url url
                     :description (or (first text) "")
                     :gid gid})
      (db/execute!))
  (send/text gid "umme ..."))

(defn link-del [gid url]
  (debug ["link-del" gid])
  (-> {:url url :gid gid}
      (links/delete-one-by-url)
      (db/execute-one!))
  (send/links gid
              "Eliminato:"
              [{:url url :text "~~ ## ~~"}]))

(defn link-search [gid text]
  (debug ["link-search" gid])
  (let [results (-> {:text text :gid gid}
                    (links/search)
                    db/execute!)]
    (send/links gid
                "Link trovati:"
                (if (empty? results)
                  [{:url (str "https://lmgtfy.com/?q=" text "&pp=1&s=d" "&s=l")
                    :text "🖕 LMGIFY"}]
                  results))))

(defn diceroll [gid text]
  (debug ["diceroll" text gid])
  (let [results (->> text
                     (dicer/roll)
                     (dicer/as-emoji)
                     (string/join ", "))]
    (when results
      (send/text gid results))))

(defn ricorda [data parsed-text]
  (debug ["ricorda" parsed-text])
  (match [data parsed-text]
    ;; photos
    [{:photo photo-sizes :chat {:id gid}}
     [_ [:text text]]]
    (do
      (doseq [item photo-sizes]
        (-> (media/insert {:file-id (:file_id item)
                           :type "photo"
                           :text text
                           :gid gid})
            (db/execute-one!)))
      (send/text gid "Uh una nuova __russacchiotta__?"))

    ;; videos
    [{:video video :chat {:id gid}}
     [_ [:text text]]]
    (do
      (-> (media/insert {:file-id (:file_id video)
                         :type "video"
                         :text text
                         :gid gid})
          db/execute-one!)
      (send/text gid "Uh un nuovo __video__\\!"))

    ;; Replies using /r foo bar
    [{:reply_to_message message :chat {:id gid}}
     [_ [:text text]]]
    (ricorda message parsed-text)

    :else
    (send/text (get-in data [:chat :id]) "non ho capito ...")))

(defn yell-callout
  ([gid co text]
   (debug ["yell-callout" gid co])
   (when-let [c (-> (callouts/one-by-callout {:callout co :gid gid})
                    (db/execute-one!))]
     (send/text gid (->> text
                         (string/replace (:text c) "%")
                         (message/escape)))))
  ([gid co]
   (yell-callout gid co "")))

(defn create-abraxas
  [gid abraxas kind]
  (debug ["create-abraxas" gid])
  (if (or (= kind "photo") (= kind "video"))
    (let [results (-> (abraxoides/insert {:gid gid :abraxas abraxas :kind kind}))]
      (-> {:gid gid :abraxas abraxas :kind kind}
          (abraxoides/insert)
          (db/execute-one!))
      (send/text gid "ho imparato una nuova evocazione!"))
    (send/text gid "non ho potuto evocare l'incantazione richiesta...")))

(defn delete-abraxas
  [gid abraxas]
  (debug {:fn "delete-abraxas" :gid gid})
  (-> {:gid gid :abraxas abraxas}
      (abraxoides/delete-by-abraxas)
      (db/execute!))
  (send/text gid "Ho dimenticato qualcosa!"))

(defn conjure-abraxas
  [gid abraxas]
  (debug ["conjure-abraxas" gid])
  (when-let [results (-> (abraxoides/search {:abraxas abraxas})
                         (db/execute-one!))]
    (debug "abraxas?" results)
    (when-let [item (-> (media/get-random-by-kind {:kind (:kind results) :gid gid})
                        (db/execute-one!))]
      (condp (:kind item)
             "photo" (send/photo gid (:data item) (:text item))
             "video" (send/video gid (:data item) (:text item))))))

(defn act [{{gid :id} :chat :as data} parsed-text]
  (debug ["act" parsed-text])
  (match parsed-text
    [_ [:command] [:abraxas "register"] & _]
    (register-channel data)

    [_ [:command] [:abraxas "ricorda"] & _]
    (ricorda data parsed-text)

    [_ [:command] [:abraxas "paris"] & _]
    (help gid)

    [_ [:command] [:abraxas "grumpy"] & _]
    (grumpyness gid)

    [_ [:command] [:abraxas (:or "l" "link" "nota")] [:add] [:url url] & text]
    (link-add gid url text)

    [_ [:command] [:abraxas (:or "l" "link" "nota")] [:del] [:url url] & _]
    (link-del gid url)

    [_ [:command] [:abraxas (:or "l" "link" "nota")] [:text text]]
    (link-search gid text)

    [_ [:command] [:abraxas (:or "d" "d20" "dice" "r" "roll")] [:text text]]
    (diceroll gid text)

    [_ [:command] [:abraxas "abraxas"] [:add] [:text text]]
    (let [[abraxas kind] (string/split text " ")]
      (create-abraxas gid abraxas kind))

    [_ [:command] [:abraxas "abraxas"] [:del] [:text abraxas]]
    (delete-abraxas gid abraxas)

    [_ [:callout] [:abraxas abx] [:text text]] (yell-callout gid abx text)
    [_ [:callout] [:abraxas abx]]              (yell-callout gid abx)

    [_ [:abraxas abx] & _] (conjure-abraxas gid abx)

    :else (do (debug "act -> no match")
              nil)))
