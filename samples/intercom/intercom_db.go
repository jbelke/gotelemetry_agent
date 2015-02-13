package plugin

import (
	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	_ "github.com/go-sql-driver/mysql"
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

	if !db.HasTable(intercomCompany{}) {
		db.CreateTable(intercomCompany{})
	}

	if !db.HasTable(intercomSegment{}) {
		db.CreateTable(intercomSegment{})
	}

	if !db.HasTable(intercomLocation{}) {
		db.CreateTable(intercomLocation{})
	}

	if !db.HasTable(intercomSocialProfile{}) {
		db.CreateTable(intercomSocialProfile{})
	}

	if !db.HasTable(intercomTag{}) {
		db.CreateTable(intercomTag{})

	}

	if !db.HasTable(intercomUser{}) {
		db.CreateTable(intercomUser{})
	}

	return nil
}
