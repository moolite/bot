package db

import (
	"database/sql"

	"github.com/rs/zerolog/log"
)

func CreateTables(dbc *sql.DB) (err error) {
	tx, err := dbc.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	_, err = tx.Exec(groupsCreateTable)
	if err != nil {
		log.Error().Err(err).Msg("error creating table `groups`")
		return err
	}
	_, err = tx.Exec(abraxoidexTable)
	if err != nil {
		log.Error().Err(err).Msg("error creating table `abraxoides`")
		return err
	}
	_, err = tx.Exec(calloutsCreateTable)
	if err != nil {
		log.Error().Err(err).Msg("error creating table `callouts`")
		return err
	}
	_, err = tx.Exec(linksCreateTable)
	if err != nil {
		log.Error().Err(err).Msg("error creating table `links`")
		return err
	}
	_, err = tx.Exec(mediaCreateTable)
	if err != nil {
		log.Error().Err(err).Msg("error creating table `media`")
		return err
	}

	return err
}
