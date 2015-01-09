package plugin

import (
	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"github.com/jinzhu/gorm"
	"github.com/telemetryapp/gotelemetry_agent/agent/job"
)

type intercomFetchIterator func(target *IntercomResponse) error

func (p *IntercomPlugin) fetchAllPages(endpoint string, iterator intercomFetchIterator) error {
	url := endpoint

	for {
		target := &IntercomResponse{}

		if err := p.performRequest(url, target); err != nil {
			return err
		}

		if err := iterator(target); err != nil {
			return err
		}

		if target.Pages.Next == nil {
			return nil
		}

		url = *target.Pages.Next
	}
}

func (p *IntercomPlugin) fetchUsers(job *job.Job) {
	job.Log("Updating all Intercom users...")

	db, err := gorm.Open("sqlite3", p.DBPath)

	if err != nil {
		job.ReportError(err)
		return
	}

	defer db.Close()

	iterator := func(r *IntercomResponse) error {
		for _, u := range r.Users {
			u.Segments = u.Segment.Segments

			for index, segment := range u.Segments {
				s := IntercomSegment{IntercomId: segment.IntercomId}
				db.FirstOrCreate(&s, s)

				u.Segments[index].Id = s.Id

				db.Save(&u.Segments[index])
			}

			u.Companies = u.Company.Companies

			for index, company := range u.Companies {
				c := IntercomCompany{IntercomId: company.IntercomId}

				db.FirstOrCreate(&c, c)

				u.Companies[index].Id = c.Id

				db.Save(&u.Companies[index])
			}

			u.Tags = u.Tag.Tags

			for index, tag := range u.Tags {
				t := IntercomTag{IntercomId: tag.IntercomId}

				db.FirstOrCreate(&t, t)

				u.Tags[index].Id = t.Id

				db.Save(&u.Tags[index])
			}

			for index, profile := range u.SocialProfiles {
				p := IntercomSocialProfile{IntercomId: profile.IntercomId}

				db.FirstOrCreate(&p, p)

				u.SocialProfiles[index].Id = p.Id

				db.Save(&u.SocialProfiles[index])
			}

			db.FirstOrCreate(&u.Location, u.Location)
			db.Save(&u.Location)

			db.FirstOrCreate(&u, IntercomUser{IntercomId: u.IntercomId})
			db.Save(&u)

			if db.Error != nil {
				return db.Error
			}
		}

		return nil
	}

	if err := p.fetchAllPages("users", iterator); err != nil {
		job.ReportError(err)
	}
}

func (p *IntercomPlugin) fetchTags(job *job.Job) {
	job.Log("Updating all Intercom tags...")

	db, err := gorm.Open("sqlite3", p.DBPath)

	if err != nil {
		job.ReportError(err)
		return
	}

	defer db.Close()

	iterator := func(r *IntercomResponse) error {
		for _, tag := range r.Tags {
			t := IntercomTag{IntercomId: tag.IntercomId}

			db.FirstOrCreate(&t, t)

			tag.Id = t.Id

			db.Save(&tag)

			if db.Error != nil {
				return db.Error
			}
		}

		return nil
	}

	if err := p.fetchAllPages("tags", iterator); err != nil {
		job.ReportError(err)
	}
}

func (p *IntercomPlugin) fetchSegments(job *job.Job) {
	job.Log("Updating all Intercom segments...")

	db, err := gorm.Open("sqlite3", p.DBPath)

	if err != nil {
		job.ReportError(err)
		return
	}

	defer db.Close()

	iterator := func(r *IntercomResponse) error {
		for _, segment := range r.Segments {
			s := IntercomSegment{IntercomId: segment.IntercomId}

			db.FirstOrCreate(&s, s)

			segment.Id = s.Id

			db.Save(&segment)

			if db.Error != nil {
				return db.Error
			}
		}

		return nil
	}

	if err := p.fetchAllPages("segments", iterator); err != nil {
		job.ReportError(err)
	}
}

func (p *IntercomPlugin) fetchCompanies(job *job.Job) {
	job.Log("Updating all Intercom companies...")

	db, err := gorm.Open("sqlite3", p.DBPath)

	if err != nil {
		job.ReportError(err)
		return
	}

	defer db.Close()

	iterator := func(r *IntercomResponse) error {
		for _, company := range r.Companies {
			c := IntercomCompany{IntercomId: company.IntercomId}

			db.FirstOrCreate(&c, c)

			company.Id = c.Id

			db.Save(&company)

			if db.Error != nil {
				return db.Error
			}
		}

		return nil
	}

	if err := p.fetchAllPages("companies", iterator); err != nil {
		job.ReportError(err)
	}
}
