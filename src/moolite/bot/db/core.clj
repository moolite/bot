(ns moolite.bot.db.core)

(defn- get-random-mad [{columns :columns table :table gid :gid}]
  {:select columns :from table
   :where [:and
           [:= :gid gid]
           [:or
            [:= :rowid [[:% [:abs [:random]]
                         {:select [[[:+ [[:max :rowid]] :1]]]
                          :from table}]]]
            [:= :rowid {:select [[[:max :rowid]]]
                        :from table}]]]})

(defn get-random [{columns :columns table :table gid :gid}]
  {:select columns
   :from table
   :where [:= :gid gid]
   :limit :1
   :offset [:%
            [:abs [:random]]
            [:max
             {:select [[:count :*]]
              :from table}
             :1]]})

(defn get-random-where [{where :where columns :columns table :table gid :gid}]
  (-> {:columns columns
       :table table
       :gid gid}
      (assoc :where [:and [:= :gid gid] where])))
