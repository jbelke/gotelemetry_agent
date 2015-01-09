package plugin

type intercomResponsePage struct {
	Next *string `json:"next"`
}

type intercomCompany struct {
	Id           int     `json:"-"`
	IntercomId   string  `json:"id"`
	InternalId   string  `json:"company_id"`
	Name         string  `json:"name"`
	UserCount    int     `json:"user_count"`
	MonthlySpend float64 `json:"monthly_spend"`
}

type intercomCompanyContainer struct {
	Companies []intercomCompany `json:"companies" gorm:"many2many:intercom_company_companies"`
}

type intercomLocation struct {
	Id       int    `json:"-"`
	City     string `json:"city_name"`
	Country  string `json:"country_code"`
	Region   string `json:"region_name"`
	Timezone string `json:"timezone"`
}

type intercomSegment struct {
	Id         int    `json:"-"`
	IntercomId string `json:"id"`
	Name       string `json:"name"`
}

type intercomSegmentContainer struct {
	Segments []intercomSegment `json:"segments"`
}

type intercomTag struct {
	Id         int    `json:"-"`
	IntercomId string `json:"id"`
	Name       string `json:"name"`
}

type intercomTagContainer struct {
	Id   int           `json:"-"`
	Tags []intercomTag `json:"tags"`
}

type intercomSocialProfile struct {
	Id         int    `json:"-"`
	IntercomId string `json:"id"`
	URL        string `json:"url"`
	Username   string `json:"username"`
	Name       string `json:"name"`
}

type intercomUser struct {
	Id                     int                      `json:"-"`
	IntercomId             string                   `json:"id"`
	InternalId             string                   `json:"user_id"`
	Company                intercomCompanyContainer `json:"companies" sql:"-" gorm:"-"`
	Companies              []intercomCompany        `json:"-" gorm:"many2many:intercom_user_companies"`
	CreatedAt              int                      `json:"created_at"`
	LastRequestAt          int                      `json:"last_request_at"`
	Email                  string                   `json:"email"`
	Location               intercomLocation         `json:"location_data"`
	LocationId             int                      ``
	Name                   string                   `json:"name"`
	Segment                intercomSegmentContainer `json:"segments" sql:"-" gorm:"-"`
	Segments               []intercomSegment        `json:"-" gorm:"many2many:intercom_user_segments"`
	SocialProfiles         []intercomSocialProfile  `json:"social_profile" gorm:"many2many:intercom_user_social_profiles"`
	SessionCount           int                      `json:"session_count"`
	Tag                    intercomTagContainer     `json:"tags" sql:"-" gorm:"-"`
	Tags                   []intercomTag            `json:"-" gorm:"many2many:intercom_user_tag"`
	UnsubscribedFromEmails bool                     `json:"unsubscribed_from_emails"`
}

type intercomResponse struct {
	Pages      intercomResponsePage `json:"pages"`
	TotalCount int                  `json:"total_count"`
	Type       string               `json:"type"`
	Users      []intercomUser       `json:"users"`
	Tags       []intercomTag        `json:"tags"`
	Segments   []intercomSegment    `json:"segments"`
	Companies  []intercomCompany    `json:"companies"`
}
