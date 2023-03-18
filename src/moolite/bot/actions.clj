;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns moolite.bot.actions
  (:require [clojure.string :as string]
            [clojure.core.match :refer [match]]
            [taoensso.timbre :as timbre :refer [info debug]]
            [moolite.bot.send :as send]
            [moolite.bot.dicer :as dicer]
            [moolite.bot.db :as db]
            [moolite.bot.db.groups :as groups]
            [moolite.bot.db.callouts :as callouts]
            [moolite.bot.db.stats :as stats]
            [moolite.bot.db.links :as links]
            [moolite.bot.db.media :as media]
            [moolite.bot.db.abraxoides :as abraxoides]))

(defn register-channel [{{gid :id title :title} :chat}]
  (debug {:fn "register-channel" :gid gid :title title})
  (let [chan (-> (groups/insert {:gid gid :title title})
                 (db/execute-one!))]
    (send/text gid (printf "channel registered with id %d and title '%s'" (:gid chan) (:title chan)))))

(defn help [gid]
  (debug {:fn "help" :gid gid})
  (let [callout (->> (callouts/all-keywords {:gid gid})
                     (db/execute-one!)
                     (map :name)
                     (map #(str "- " %))
                     (string/join "\n"))]
    (send/text gid callout)))

(defn grumpyness [gid]
  (debug {:fn "grumpyness" :gid gid})
  (let [grumpyness (-> (stats/all {:gid gid})
                       (db/execute-one!))]
    (send/text gid grumpyness)))

(defn link-add [gid url & text]
  (debug {:fn "link-add" :gid gid})
  (-> (links/insert {:url url
                     :description (or (first text) "")
                     :gid gid})
      (db/execute!))
  (send/text gid "umme ..."))

(defn link-del [gid url]
  (debug {:fn "link-del" :gid gid})
  (-> {:url url :gid gid}
      (links/delete-one-by-url)
      (db/execute-one!))
  (send/links gid
              "Eliminato:"
              [{:url url :text "~~ ## ~~"}]))

(defn link-search [gid text]
  (debug {:fn "link-search" :gid gid})
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
  (debug {:fn "diceroll" :gid gid})
  (let [results (->> text
                     dicer/roll
                     dicer/as-emoji
                     (string/join ", "))]
    (when results
      (send/dice gid results))))

(defn ricorda [data parsed-text]
  (debug {:fn "ricorda" :text parsed-text})
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
      (send/text gid "Uh una nuova _russacchiotta_?"))

    ;; videos
    [{:video video :chat {:id gid}}
     [_ [:text text]]]
    (do
      (-> (media/insert {:file-id (:file_id video)
                         :type "video"
                         :text text
                         :gid gid})
          db/execute-one!)
      (send/text gid "Uh un nuovo _video_!"))

    ;; Replies using /r foo bar
    [{:reply_to_message message :chat {:id gid}}
     [_ [:text text]]]
    (ricorda message parsed-text)))

(defn yell-callout
  ([gid co text]
   (debug {:fn "yell-callout" :gid gid :co co})
   (when-let [c (-> (callouts/one-by-callout {:callout co :gid gid})
                    (db/execute-one!))]
     (send/text gid (string/replace (:text c) "%" text))))
  ([gid co]
   (yell-callout gid co "")))

(defn conjure-abraxas
  [gid abx]
  (debug {:fn "conjure-abraxas" :gid gid})
  (let [results (-> (abraxoides/search abx)
                    (db/execute-one!))
        item (-> (media/get-random-by-kind {:kind (:kind results) :gid gid})
                 (db/execute-one!))]
    (when item
      (condp (:kind item)
             "photo" (send/photo gid (:media_id item) (:text item))
             "video" (send/photo gid (:media_id item) (:text item))))))

(defn act [{{gid :id} :chat :as data} parsed-text]
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

    [_ [:command] [:abraxas (:or "d" "d20" "dice")] [:text text]]
    (diceroll gid text)

    [_ [:callout] [:abraxas abx] [:text text]] (yell-callout gid abx text)
    [_ [:callout] [:abraxas abx]]              (yell-callout gid abx)

    [_ [:abraxas abx] & _] (conjure-abraxas gid abx)

    :else (do (debug "act -> no match")
              nil)))
