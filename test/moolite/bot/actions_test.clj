(ns moolite.bot.actions-test
  (:require [clojure.pprint :refer [pprint]]
            [moolite.bot.actions :as a]
            [moolite.bot.db :as db]
            [clojure.core.match :refer [match]]
            [clojure.test :refer [deftest]]))

(def gid 1)
(def url "https://example.com")

(defn should-be [kind data]
  (match kind
    :message (assert true)))

(deftest test-register-channel
  (let [register-result (a/register-channel
                         {:chat {:id gid :title "foo"}})
        find-result (db/channels-search {:gid gid})]
    (pprint register-result)
    (pprint find-result)))

(deftest test-help)

(deftest test-link
  ;; link-add
  (let [url "https://example.com"
        response (a/link-add gid url "Example")
        result (db/links-get-by-url url)]
    (pprint response)
    (pprint result))

  ;; link-del
  (let [add-result (a/link-add gid url "Example")
        del-result (a/link-del gid url)
        search-result (db/links-get-by-url url)]
    (pprint search-result)))

(deftest test-diceroll
  (let [results (a/diceroll gid "1d20")]
    (should-be :message results)))

(deftest test-ricorda)
(deftest test-ricorda-reply)

(deftest test-yell-callout
  (let [db-results (db/abraxoides-insert {:abraxas "foo" :kind ""})
        results (a/yell-callout gid "foo")]))

(deftest test-conjure-abraxas)

(deftest test-act)
