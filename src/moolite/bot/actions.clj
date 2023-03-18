;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at http://mozilla.org/MPL/2.0/.
(ns moolite.bot.actions
  (:require [clojure.string :as string]
            [moolite.bot.send :as send]
            [moolite.bot.dicer :as dicer]
            [moolite.bot.db :as db]
            [moolite.bot.db.groups :as groups]
            [moolite.bot.db.callouts :as callouts]
            [moolite.bot.db.stats :as stats]
            [moolite.bot.db.links :as links]
            [moolite.bot.db.media :as media]
            [moolite.bot.db.abraxoides :as abraxoides]
            [clojure.core.match :refer [match]]))

(defn register-channel [{{gid :id title :title} :chat}]
  (let [chan (-> (groups/insert {:gid gid :title title})
                 (db/execute-one!))]
    (send/text gid (printf "channel registered with id %d and title '%s'" (:gid chan) (:title chan)))))

(defn help [gid]
  (let [callout (->> (callouts/all-keywords {:gid gid})
                     (db/execute-one!)
                     (map :name)
                     (map #(str "- " %))
                     (string/join "\n"))]
    (send/text gid callout)))

(defn grumpyness [gid]
  (let [grumpyness (-> (stats/all {:gid gid})
                       (db/execute-one!))]
    (send/text gid grumpyness)))

(defn link-add [gid url & text]
  (-> (links/insert {:url url
                     :text (or (first text) "")
                     :gid gid})
      (db/execute!)))

(defn link-del [gid url]
  (let [deleted (-> {:url url :gid gid}
                    (links/delete-one-by-url)
                    db/execute-one!)]
    (send/text gid (str "Eliminati: \n" (string/join "\n" (map :url deleted))))))

(defn link-search [gid text]
  (let [results (-> (links/search {:text text :gid gid})
                    db/execute!)]
    (send/links gid
                "Link trovati:"
                (if (empty? results)
                  [{:url (str "https://lmgtfy.com/?q=" text "&pp=1&s=d" "&s=l")
                    :text "🖕 LMGIFY"}]
                  results))))

(defn diceroll [gid text]
  (let [results (->> text
                     dicer/roll
                     dicer/as-emoji
                     (string/join ", "))]
    (when results
      (send/dice gid results))))

(defn ricorda [data parsed-text]
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
   (when-let [c (-> (callouts/one-by-callout {:callout co :gid gid})
                    (db/execute-one!))]
     (send/text gid (string/replace (:text c) "%" text))))
  ([gid co]
   (yell-callout gid co "")))

(defn conjure-abraxas
  [gid abx]
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
    [[[:command] [:abraxas "register"] & _]]
    (register-channel data)

    [[[:command] [:abraxas "ricorda"] & _]]
    (ricorda data parsed-text)

    [[[:command] [:abraxas "paris"] & _]]
    (help gid)

    [[[:command] [:abraxas "grumpy"] & _]]
    (grumpyness gid)

    [[[:command] [:abraxas (:or "l" "link" "nota")] [:add] [:url url] & text]]
    (link-add gid url text)

    [[[:command] [:abraxas (:or "l" "link" "nota")] [:del] [:url url] & _]]
    (link-del gid url)

    [[[:command] [:abraxas (:or "l" "link" "nota")] [:text text]]]
    (link-search gid text)

    [[[:command] [:abraxas (:or "d" "d20" "dice")] [:text text]]]
    (diceroll gid text)

    [[[:callout] [:abraxas abx] [:text text]]] (yell-callout gid abx text)
    [[[:callout] [:abraxas abx]]]              (yell-callout gid abx)

    [[[:abraxas abx] & _]] (conjure-abraxas gid abx)

    :else ""))
