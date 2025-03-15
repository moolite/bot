package db

import (
	"context"
)

var channelStatsTable string = "channel_stats"

type ChannelStats struct {
	GID    int64  `db:"gid"`
	UID    string `db:"uid"`
	Points int64  `db:"points"`
}

func SelectChannelStats(ctx context.Context, channel int64) ([]ChannelStats, error) {
	res := []ChannelStats{}
	q, err := prepareStmt(
		`SELECT gid, points, uid FROM channel_stats WHERE gid = ?`,
	)
	if err != nil {
		return res, err
	}

	return res, q.SelectContext(ctx, &res, channel)
}

func SelectChannelStatsUser(ctx context.Context, channel, user string) (*ChannelStats, error) {
	q, err := prepareStmt(
		`SELECT * FROM ` + channelStatsTable + ` WHERE gid = ? AND user = ?`,
	)
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

func InsertChannelStats(ctx context.Context, c *ChannelStats) error {
	q, err := prepareStmt(
		`INSERT INTO ` + channelStatsTable + `
				(gid,uid,points) VALUES(?,?,?)
			ON CONFLICT(gid,uid) DO UPDATE SET
				points=(points + excluded.points)
				WHERE gid=excluded.gid AND uid=excluded.uid
			RETURNING points
			`,
	)
	if err != nil {
		return err
	}
	return q.GetContext(ctx, c, c.GID, c.UID, c.Points)
}
