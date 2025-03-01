package db

import (
	"context"
)

var channelStatsTable string = "channel_stats"

type ChannelStats struct {
	GID    int64
	User   string
	Points int64
}

func SelectChannelStats(ctx context.Context, channel string) ([]*ChannelStats, error) {
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

func SelectChannelStatsUser(ctx context.Context, channel, user string) (*ChannelStats, error) {
	q, err := prepareStmt(`SELECT * FROM ` + channelStatsTable + ` WHERE gid = ? AND user = ?`)
	if err != nil {
		return nil, err
	}

	row := q.QueryRowContext(ctx, channel, user)
	res := &ChannelStats{}
	if err := row.Scan(res); err != nil {
		return nil, err
	}

	return res, nil
}

func InsertChannelStats(ctx context.Context, channel int64, user string, points int64) (*ChannelStats, error) {
	res := &ChannelStats{
		GID:    channel,
		User:   user,
		Points: points,
	}
	q, err := prepareStmt(`INSERT INTO ` + channelStatsTable + `
	(gid,uid,points)
	VALUES(?,?,?)
	ON CONFLICT(uid,ts) DO
		UPDATE SET points=(` + channelStatsTable + `.points + points)
	RETURNING *`)
	if err != nil {
		return nil, err
	}

	row := q.QueryRowContext(ctx, res.GID, res.User, res.Points)

	if err := row.Scan(res); err != nil {
		return nil, err
	}

	return res, nil
}
