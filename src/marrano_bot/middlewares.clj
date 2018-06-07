(ns marrano-bot.middlewares
  (:require [clojure.core.async :refer [go-loop <! >!! sliding-buffer chan]]))

(defn logger
  "logging middleware"
  [handler]
  (let [logging-chan (chan (sliding-buffer 100))]
    (go-loop []
      (let [req    (<! logging-chan)
            method (:request-method req)
            uri    (:uri req)]
        (println method uri))
      (recur))
    (fn [req]
      (>!! logging-chan req)
      (handler req))))
