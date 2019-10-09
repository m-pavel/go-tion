package main

import (
	"database/sql"

	"github.com/gorhill/cronexpr"
	_ "github.com/mattn/go-sqlite3"
)

// Dao of the scheduler
type Dao struct {
	db *sql.DB
}

// Schedule table
type Schedule struct {
	ID      int
	Value   string
	Enabled *bool
	Heater  *bool
	Gate    *int
	Speed   *int
	Sound   *bool
	Temp    *int
}

// New Dao
func New(db string) (*Dao, error) {
	dao := Dao{}
	var err error
	dao.db, err = sql.Open("sqlite3", db)
	if err != nil {
		return nil, err
	}
	return &dao, nil
}

// Prepare table
func (d *Dao) Prepare() error {
	_, err := d.db.Exec("CREATE TABLE SCHEDULES (SCHEDULE text NOT NULL, ENABLED int, HEATER int, SOUND int, GATE int, SPEED int, TEMP int)")
	return err
}

// Close dao
func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
		d.db = nil
	}
}

// GetSchedules from DB
func (d *Dao) GetSchedules() ([]Schedule, error) {
	stmt, err := d.db.Prepare("SELECT ROWID, SCHEDULE, ENABLED, HEATER, SOUND, GATE, SPEED, TEMP FROM SCHEDULES")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sch := make([]Schedule, 0)
	for {
		if !rows.Next() {
			break
		}
		var s Schedule
		rows.Scan(&s.ID, &s.Value, &s.Enabled, &s.Heater, &s.Sound, &s.Gate, &s.Speed, &s.Temp)
		if err != nil {
			return nil, err
		}
		sch = append(sch, s)
	}
	return sch, err
}

// Add schedule
func (d *Dao) Add(schedule string, enabled *bool, heater *bool, sound *bool, gate *int, speed *int, temp *int) error {
	_, err := cronexpr.Parse(schedule)
	if err != nil {
		return err
	}
	_, err = d.db.Exec("INSERT INTO SCHEDULES (SCHEDULE, ENABLED, HEATER, SOUND, GATE, SPEED, TEMP) VALUES (?, ?, ?, ?, ?, ?, ?)",
		schedule, enabled, heater, sound, gate, speed, temp)
	return err
}

// Delete entry
func (d *Dao) Delete(id int) error {
	_, err := d.db.Exec("DELETE FROM SCHEDULES WHERE ROWID=?", id)
	return err
}

// UpdateHeater heater mode
func (d *Dao) UpdateHeater(heater bool) error {
	_, err := d.db.Exec("UPDATE SCHEDULES SET HEATER=?", heater)
	return err
}
