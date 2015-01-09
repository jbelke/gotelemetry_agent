package plugin

import (
	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"github.com/jinzhu/gorm"
	"github.com/telemetryapp/gotelemetry_agent/agent/job"
)

func (p *IntercomPlugin) setupDatabase(job *job.Job) error {
	db, err := gorm.Open("sqlite3", p.DBPath)

	defer db.Close()

	if err != nil {
		return err
	}

	// db.LogMode(true)

	if !db.HasTable(IntercomCompany{}) {
		db.CreateTable(IntercomCompany{})
	}

	if !db.HasTable(IntercomSegment{}) {
		db.CreateTable(IntercomSegment{})
	}

	if !db.HasTable(IntercomLocation{}) {
		db.CreateTable(IntercomLocation{})
	}

	if !db.HasTable(IntercomSocialProfile{}) {
		db.CreateTable(IntercomSocialProfile{})
	}

	if !db.HasTable(IntercomTag{}) {
		db.CreateTable(IntercomTag{})

	}

	if !db.HasTable(IntercomUser{}) {
		db.CreateTable(IntercomUser{})
	}

	return nil
}
