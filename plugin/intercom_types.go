package plugin

type IntercomResponsePage struct {
	Next *string `json:"next"`
}

type IntercomCompany struct {
	Id           int     `json:"-"`
	IntercomId   string  `json:"id"`
	InternalId   string  `json:"company_id"`
	Name         string  `json:"name"`
	UserCount    int     `json:"user_count"`
	MonthlySpend float64 `json:"monthly_spend"`
}

type IntercomCompanyContainer struct {
	Companies []IntercomCompany `json:"companies" gorm:"many2many:intercom_company_companies"`
}

type IntercomLocation struct {
	Id       int    `json:"-"`
	City     string `json:"city_name"`
	Country  string `json:"country_code"`
	Region   string `json:"region_name"`
	Timezone string `json:"timezone"`
}

type IntercomSegment struct {
	Id         int    `json:"-"`
	IntercomId string `json:"id"`
	Name       string `json:"name"`
}

type IntercomSegmentContainer struct {
	Segments []IntercomSegment `json:"segments"`
}

type IntercomTag struct {
	Id         int    `json:"-"`
	IntercomId string `json:"id"`
	Name       string `json:"name"`
}

type IntercomTagContainer struct {
	Id   int           `json:"-"`
	Tags []IntercomTag `json:"tags"`
}

type IntercomSocialProfile struct {
	Id         int    `json:"-"`
	IntercomId string `json:"id"`
	URL        string `json:"url"`
	Username   string `json:"username"`
	Name       string `json:"name"`
}

type IntercomUser struct {
	Id                     int                      `json:"-"`
	IntercomId             string                   `json:"id"`
	InternalId             string                   `json:"user_id"`
	Company                IntercomCompanyContainer `json:"companies" sql:"-" gorm:"-"`
	Companies              []IntercomCompany        `json:"-" gorm:"many2many:intercom_user_companies"`
	CreatedAt              int                      `json:"created_at"`
	LastRequestAt          int                      `json:"last_request_at"`
	Email                  string                   `json:"email"`
	Location               IntercomLocation         `json:"location_data"`
	LocationId             int                      ``
	Name                   string                   `json:"name"`
	Segment                IntercomSegmentContainer `json:"segments" sql:"-" gorm:"-"`
	Segments               []IntercomSegment        `json:"-" gorm:"many2many:intercom_user_segments"`
	SocialProfiles         []IntercomSocialProfile  `json:"social_profile" gorm:"many2many:intercom_user_social_profiles"`
	SessionCount           int                      `json:"session_count"`
	Tag                    IntercomTagContainer     `json:"tags" sql:"-" gorm:"-"`
	Tags                   []IntercomTag            `json:"-" gorm:"many2many:intercom_user_tag"`
	UnsubscribedFromEmails bool                     `json:"unsubscribed_from_emails"`
}

type IntercomResponse struct {
	Pages      IntercomResponsePage `json:"pages"`
	TotalCount int                  `json:"total_count"`
	Type       string               `json:"type"`
	Users      []IntercomUser       `json:"users"`
	Tags       []IntercomTag        `json:"tags"`
	Segments   []IntercomSegment    `json:"segments"`
	Companies  []IntercomCompany    `json:"companies"`
}
