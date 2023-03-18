(ns moolite.bot.actions-test
  (:require [clojure.pprint :refer [pprint]]
            [clojure.core.match :refer [match]]
            [clojure.test :refer [deftest is]]
            [moolite.bot.actions :as a]
            [moolite.bot.db :as db]
            [moolite.bot.db.groups :as groups]
            [moolite.bot.db.links :as links]
            [moolite.bot.db.abraxoides :as abraxoides]
            [moolite.bot.db.stats :as stats]))

(def gid 1)
(def url "https://example.com")

(defn should-be [kind data]
  (match kind
    :message (assert true)))

(deftest test-register-channel
  (-> (groups/delete-one {:gid gid})
      (db/execute!))
  (let [register-result (a/register-channel
                         {:chat {:id gid :title "foo"}})
        find-result (groups/get-one {:gid gid})]
    (pprint register-result)
    (pprint find-result)))

(deftest test-help)

(deftest test-link
  ;; link-add
  (let [url "https://example.com"
        response (a/link-add gid url "Example")
        result (links/get-by-url url)]
    (pprint response)
    (pprint result))

  ;; link-del
  (let [add-result (a/link-add gid url "Example")
        del-result (a/link-del gid url)
        search-result (-> {:url url :gid gid}
                          (links/get-by-url)
                          (db/execute!))]
    (pprint search-result)))

(deftest test-diceroll
  (let [results (a/diceroll gid "1d20")]
    (should-be :message results)))

(deftest test-ricorda)
(deftest test-ricorda-reply)

(deftest test-yell-callout
  (let [db-results (-> {:gid gid :abraxas "foo" :kind ""}
                       (abraxoides/insert)
                       (db/execute!))
        results (a/yell-callout gid "foo")]))

(deftest test-conjure-abraxas)

(deftest test-act)

(deftest test-stats
  (-> {:gid gid :keyword "foo"}
      (stats/delete-one)
      (db/execute-one!))
  (-> {:gid gid :keyword "foo"}
      (stats/insert)
      (db/execute-one!))
  (-> {:gid gid :keyword "foo"}
      (stats/insert)
      (db/execute-one!)
      (:stat)
      (is 2)))
