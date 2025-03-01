package db

import (
	"context"
	"time"
)

const (
	statisticsTable     string = `statistics`
	statisticsKindTable string = `statistics_kind`
)

type StatisticsKind struct {
	KindID   int64
	Name     string
	Trigger  string
	IsRegexp bool
}

type Statistics struct {
	KindID int64
	Value  int64
	Date   time.Time
}

type StatisticsJoin struct {
	Name  string
	Value int64
	Date  time.Time
}

func SelectStatisticKinds(ctx context.Context) ([]*StatisticsKind, error) {
	var results []*StatisticsKind
	q, err := prepareStmt(
		`SELECT kind_id,name,is_regexp FROM ` + statisticsKindTable,
	)
	if err != nil {
		return results, err
	}

	rows, err := q.QueryContext(ctx)
	if err != nil {
		return results, err
	}

	for rows.Next() {
		kind := &StatisticsKind{}
		err := rows.Scan(&kind.KindID, &kind.Name, &kind.IsRegexp)
		if err != nil {
			return results, err
		}

		results = append(results, kind)
	}

	return results, nil
}

func InsertStatistics(ctx context.Context, val, kind int64) error {
	q, err := prepareStmt(
		`INSERT INTO statistics SET(value,kind) VALUES(?,?)`,
	)
	if err != nil {
		return err
	}

	rows, err := q.ExecContext(ctx, val, kind)
	if err != nil {
		return err
	}

	n, err := rows.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrInsert
	}

	return nil
}

func SelectStatisticsByDateRange(ctx context.Context, timeFrom, timeTo time.Time) ([]*StatisticsJoin, error) {
	var results []*StatisticsJoin

	q, err := prepareStmt(
		`SELECT name,value,date FROM statistics
		WHERE date < ? AND date > ?
		LEFT JOIN statistics_kind USING(kind_id)`,
	)
	if err != nil {
		return results, err
	}

	rows, err := q.QueryContext(ctx, timeFrom.Unix(), timeTo.Unix())
	if err != nil {
		return results, err
	}

	for rows.Next() {
		dest := &StatisticsJoin{}
		err = rows.Scan(&dest.Value, &dest.Name, &dest.Value, &dest.Date)
		if err != nil {
			return results, err
		}

		results = append(results, dest)
	}

	return results, nil
}

func SelectStatisticsLatest(ctx context.Context) ([]*StatisticsJoin, error) {
	var results []*StatisticsJoin
	q, err := prepareStmt(
		`SELECT name,value,date FROM statistics
		LEFT JOIN statistics_kind USING(kind_id)
		WHERE date > date('now','-30 minutes')
		ORDER BY date`,
	)
	if err != nil {
		return results, err
	}

	rows, err := q.QueryContext(ctx)
	if err != nil {
		return results, err
	}

	for rows.Next() {
		dest := &StatisticsJoin{}
		err = rows.Scan(&dest.Value, &dest.Name, &dest.Value, &dest.Date)
		if err != nil {
			return results, err
		}

		results = append(results, dest)
	}

	return results, nil
}

func InsertStatisticsKind(ctx context.Context, k *StatisticsKind) (int64, error) {
	q, err := prepareStmt(
		`INSERT INTO statistics_kind (name,trigger,is_regexp) VALUES(?,?,?)`,
	)
	if err != nil {
		return -1, err
	}

	row, err := q.ExecContext(ctx, k.Name, k.Trigger, k.IsRegexp)
	if err != nil {
		return -1, err
	}

	id, err := row.LastInsertId()
	if err != nil {
		return -1, err
	}
	if id == 0 {
		return -1, err
	}

	return id, nil
}
