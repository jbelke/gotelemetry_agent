package plugin

import (
	"github.com/beefsack/go-rate"
	"github.com/telemetryapp/gotelemetry_agent/agent/job"
	"strconv"
	"time"
)

// Function init() registers this plugin with the Plugin Manager.
// The plugin provides an IntercomPluginFactory that the manager calls whenever it needs
// to create a new job
func init() {
	job.RegisterPlugin("com.telemetryapp.intercom", IntercomPluginFactory)
}

// Func IntercomPluginFactory generates a blank plugin instance for the
// `com.telemetryapp.intercom` plugin
func IntercomPluginFactory() job.PluginInstance {
	return &IntercomPlugin{
		PluginHelper: job.NewPluginHelper(),
	}
}

// Struct IntercomPlugin allows pulling and collating information from an Intercom.io
// account (https://intercom.io).
//
// The plugin does not populate Telemetry flows directly; instead, it stores everything
// into a local SQLite database that you specify through the configuration file.
//
// The database contains these tables:
//
// - intercom_organizations      - a list of organizations registered in your Intercom account
// - intercom_users              - a list of users registered in your Intercom account
// - intercom_segments           - a list of segments registered in your Intercom account
// - intercom_tags               - a list of tags registered in your Intercom account
// - intercom_social_profiles    - a list of social profiles for each user
//
// Join tables are also provided that link users to companies, tags, segments, and social profiles
//
// For information on configuration parameters, check IntercomPlugin.Init()
type IntercomPlugin struct {
	*job.PluginHelper
	AppID   string
	APIKey  string
	DBPath  string
	Limiter *rate.RateLimiter
}

// Function Init initializes the plugin.
//
// Required configuration parameters are:
//
// - application_id - Your Intercom Application ID
// - api_key        - Your (Read only) Intercom API Key
// - db_path        - A location for the output SQLite database.
//                    Note that the agent must have create and write access to this location
//
// Your Intercom Application ID and API Key can be retrieved from Intercom by choosing Company Settings / API Keys
func (p *IntercomPlugin) Init(job *job.Job) error {
	config := job.Config()

	p.AppID = config["application_id"].(string)
	p.APIKey = config["api_key"].(string)
	p.DBPath = config["db_path"].(string)

	refresh := time.Duration(config["refresh"].(int)) * time.Second

	// Make a sample request to determine rate limiting

	p.Limiter = rate.New(100, 120*time.Second)

	headers, err := p.performRequestAndGetHeaders("tags", nil)

	if err != nil {
		return err
	}

	limit, err := strconv.Atoi(headers.Get("X-Ratelimit-Limit"))

	if err != nil {
		return err
	}

	p.Limiter = rate.New(limit, 120*time.Second)

	job.Logf("Intercom API limit is %d every 2 minutes", limit)

	// Setup database

	if err := p.setupDatabase(job); err != nil {
		return err
	}

	p.PluginHelper.AddTaskWithClosure(p.performAllTasks, refresh)

	return nil
}

func (p *IntercomPlugin) performAllTasks(job *job.Job) {
	job.Log("Starting Intercom plugin...")

	// p.fetchTags(job)
	// p.fetchSegments(job)
	// p.fetchCompanies(job)
	// p.fetchUsers(job)

	job.Log("Intercom plugin done.")

	job.PerformSubtasks()
}
