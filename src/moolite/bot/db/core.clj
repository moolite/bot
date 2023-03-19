(ns moolite.bot.db.core)

(defn- __get-random [{:keys [columns table gid where]}]
  {:select columns :from table
   :where [:and
           [:and
            [:= :gid gid]
            where]
           [:or
            [:= :rowid [[:% [:abs [:random]]
                         {:select [[[:+ [[:max :rowid]] :1]]]
                          :from table}]]]
            [:= :rowid {:select [[[:max :rowid]]]
                        :from table}]]]})

(defn get-random [{:keys [columns table where]}]
  {:select columns
   :from table
   :where where
   :limit :1
   :offset [:%
            [:abs [:random]]
            [:max
             {:select [[:count :*]]
              :from table}
             :1]]})
