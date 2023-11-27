package db

import (
	"context"
	"time"
)

var channelStatsTable string = "channel_stats"

type ChannelStats struct {
	GID    string
	TS     time.Time
	User   string
	Points int64
}

func SelectChannelStatsTimeSeries(ctx context.Context, channel string) ([]*ChannelStats, error) {
	var res []*ChannelStats
	q := `SELECT gid, SUM(points), DISTINCT(uid) WHERE gid = ?`
	rows, err := dbc.QueryContext(ctx, q, channel)
	if err != nil {
		return res, err
	}

	for rows.Next() {
		d := &ChannelStats{}
		if err := rows.Scan(&d.GID, &d.Points, &d.User); err != nil {
			return res, err
		}

		res = append(res, d)
	}
	return res, nil
}

func InsertChannelStats(ctx context.Context, channel, user string, points int64) error {
	q, err := prepareStmt(`INSERT INTO ` + channelStatsTable + `
	(gid,uid,points)
	VALUES(?,?,?)
	ON CONFLICT(uid,ts) DO
		UPDATE SET points=(` + channelStatsTable + `.points + points)`)
	if err != nil {
		return err
	}

	res, err := q.ExecContext(ctx, channel, user, points)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrInsert
	}

	return nil
}
