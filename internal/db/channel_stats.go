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

type ChannelStatsStats struct {
	GID   int64 `db:"gid"`
	Min   int   `db:"min"`
	Max   int   `db:"max"`
	Sum   int   `db:"sum"`
	Avg   int   `db:"avg"`
	Count int   `db:"count"`
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

func SelectChannelStatsStats(ctx context.Context, gid int64) (*ChannelStatsStats, error) {
	res := &ChannelStatsStats{}

	q, err := prepareStmt(
		`SELECT
			gid,
			MIN(points) AS min,
			MAX(points) AS max,
			SUM(points) AS sum,
			CAST(ROUND(AVG(points)) AS INTEGER) AS avg
			COUNT(points) AS count,
		FROM channel_stats
		WHERE gid = ?
		`)
	if err != nil {
		return res, err
	}

	err = q.QueryRowContext(ctx, gid).Scan(res)
	return res, err
}

func SelectChannelStatsUser(ctx context.Context, channel int64, user string) (*ChannelStats, error) {
	res := &ChannelStats{}
	q, err := prepareStmt(
		`SELECT * FROM ` + channelStatsTable + ` WHERE gid = ? AND user = ?`,
	)

	if err != nil {
		return res, err
	}
	err = q.QueryRowContext(ctx, channel, user).Scan(res)
	return res, err
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
