# Telemetry Agent

The Telemetry Agent simplifies the process of creating daemon processes that feed data into one or more [Telemetry](http://telemetryapp.com) flows.

Typical use-case scenarios include:

  - Feeding data from existing infrastructure (e.g.: a MySQL database) to a board
  - Automatically creating boards for your customers
  - Interfacing third-party APIs with Telemetry

The Agent is written in Go and runs fine on most Linux distros, OS X, and Windows. It is designed to run on your infrastructure, and its only requirement is that it be able to reach the Telemetry API endpoint (https://api.telemetryapp.com) on port 443 via HTTPS. It can therefore happily live behind firewalls without posing a security risk.

## Quickstart

If you want to use the Agent in conjunction with an existing plugin, all you need to do is write a configuration file, compile the Agent for your platform, and deploy both to a location that has access to the resources from which you want to pull data.

```yaml
accounts: 
  - api_key: <your key>
    submission_interval: 1
    jobs:

      - id: Random board
        plugin: random
        config:
          board:
            name: Random Board Test 3
            prefix: random_3
```

The configuration file is written in [YAML](http://www.yaml.org), a text-based markup language that requires little in the way of formatting.

The topmost object in the configuration file is an object that contains an array of account entries. Each account entry, in turn, provides the API key of the Telemetry account to which you want to write, as well as the interval at which updates are sent to the Telemetry API.

In addition, each account entry contains a list of jobs, which tell the Agent exactly what you want it to do. A job contains:

- A unique `id` that identifies it. This must be unique across your entire configuration file.

- A `submission_interval` that determines how often data is sent to the Telemetry API. The Agent coalesces updates and sends them at this interval in order to reduce your API usage.

- A `config` hash that contains the configuration for the plugin. The contents of this hash depend on which plugin you use

### Running the Agent

The Agent should compile without problems on any platform that is supported by [Go](http://golang.org). Once compiled, you can deploy the Agent's executable and point it to the config file thusly:

`agent -config /var/telemetry/agent_config.yaml`

## Plugins

Plugins are at the heart of the Agent's functionality: They provide the glue that bridges a data source to the Telemetry API.

The Agent does not currently come with any built-in plugins—although it eventually will. In the meantime, writing your own plugin is a straightforward affair that, in most cases, only requires a few lines of code whose complexity depends on how hard extracting data from your source is.

### Reasons for having plugins

The Agent's design is motivated by two goals. The first is to make it simple for third parties to integrate their services with Telemetry. For example, you can very easily write a plugin that creates an entire board and populates it with a set of data; you can then replicate that board across any number of your clients with nothing more than a few configuration lines. Similarly, you can write a plugin that “knows” how to pull data from your API and feed it to a Telemetry flow, thus making it trivial for your customers to add it to their custom boards without having to figure out how to create the integration themselves.

The second is that many users of Telemetry want to display custom visualizations based on data that resides behind firewalls, or that would otherwise hard for Telemetry to access from the outside (e.g.: because it requires a special set of credentials, and so on). Using the Agent allows you to keep your data safe and under you control until it is sent to the Telemetry API.

By implementing a simple plugin system, the Agent allows you to mix and match data sources as needed, as well as adding your own data sources without having to worry about the mechanics of running a daemon that knows how to talk to the Telemetry API.

## Writing a plugin

The Agent tries hard to stay out of your way so that you can focus on only writing the bits of functionality that are specific to your needs without having to worry about all the scaffolding and behind-the-scenes work that's required to run a high-performance daemon capable of conversing with the Telemetry API.

Even though the Agent is fully multithreaded and can easily handle hundreds, or even thousands, of data sources in a single process with minimal resource usage, the plugin infrastructure is built so as to run in a mostly synchronous way to make developing plugins as easy as possible.

Plugins are fairly easy to write—in most cases, your only scaffolding will consist of a single function and a dozen or so lines of code. You can take a look at the `/plugin/random.go` file for a simple, fully-commented example.

### Getting started

In most cases, you will want to start by forking the Agent from Github and creating a file to hold your plugin code. If you intend to contribute the plugin back to the community so that others may be able to use it, you will want to place it in the `/plugin` directory. If, on the other hand, you are creating a custom plugin that is not meant for sharing, you can drop it in the `plugin/custom` directory instead, where git will safely ignore it.

Each plugin is responsible for registering itself with the Agent when it starts so that it can be instantiated at runtime. This is typically done by adding an `init` function to your plugin file, and have it call the Agent's plugin manager thusly:

```go
func init() {
  job.RegisterPlugin("com.telemetryapp.random", RandomPluginFactory)
}
```

As you can see, [`RegisterPlugin`](http://godoc.org/github.com/telemetryapp/gotelemetry_agent/agent/job#RegisterPlugin) takes a name and a factory function. Names, which are used to identify the plugin in config files, are globally unique; therefore, we recommend that you use a reverse-domain notation to avoid conflict with other providers.

The factory function is responsible for creating instances of your plugin. At runtime, the Agent will read through the config file, and create instances of your plugin every time it needs the functionality it provides:

```go
func RandomPluginFactory() job.PluginInstance {
  return &RandomPlugin{
    job.NewPluginHelper(),
  }
}
```

Note that the factory function is _not_ responsible for configuring a plugin for runtime use. Therefore, it is usually very simple.

### Using the Plugin Helper

At runtime, plugin instances are represented by structs that conform to the [`PluginInstance`](http://godoc.org/github.com/telemetryapp/gotelemetry_agent/agent/job#PluginInstance) interface.

Although plugins can be very sophisticated, in most cases they will typically perform a combination of two functions:

  - Create a board
  - Create and/or populate one or more flows on a schedule

If this sounds like what you want to do, the Agent comes with a simplified plugin class that takes care of most of the scaffolding for you, called [`PluginHelper`](http://godoc.org/github.com/telemetryapp/gotelemetry_agent/agent/job#PluginHelper).

`PluginHelper` has most of the code needed to run a plugin; you simply provide it with a list of “tasks” that it then executes at specific intervals. Each task is just a function that runs synchronously, but independently of every other task.

Start by creating a struct to represent your plugin, and use `PluginHelper` as the basis for it:

```go

import (
  "github.com/telemetryapp/gotelemetry_agent/agent/job"
)

type RandomPlugin struct {
  *job.PluginHelper
}
```

Your plugin now satisfies most of the requirements of the `PluginInstance` interface; the only thing you need is an initialization method:

```go
func (r *RandomPlugin) Init(job *job.Job) error {
}
```

### A bit about Jobs

Before we move forwrad, a quick detour to talk about jobs. A Job is the Agent's smallest unit of work. It represents the union of a set of Telemetry credentials, a plugin instance, and a set of configuration parameters.

At runtime, Jobs are represented by the [`Job`](http://godoc.org/github.com/telemetryapp/gotelemetry_agent/agent/job#Job) struct, an instance of which is passed to your plugin instances at every step of their execution.

The `Job` instance is your plugin's interface to the outside world. It allows you to make calls to the Telemetry API without having to worry about holding credentials, report runtime errors to the Agent, and even log interesting data that you want users to be able to see.

### Creating a board

Your plugins can generate and populate entire boards based on a template layout. This handy if, for example, you want to provide each of your customers with their own copy of the same board, and populate that board with custom data that varies from customer to customer.

The easiest way to do so is to first design your board in the Telemetry editor, then use the [`boarddump`](https://github.com/telemetryapp/gotelemetry/tree/master/boarddump) app that comes with the [gotelemetry](https://github.com/telemetryapp/gotelemetry) project. `boarddump` outputs a simple JSON-encoded string that you can easily embed in your plugins [in the form of a string](https://github.com/telemetryapp/gotelemetry_agent/blob/master/plugin/random.go#L60).

When your plugin is initialized, you can then ask the agent to create a board based on your template:

```go
boardName := "My Board"
boardPrefix := "custom-"

template := "<your template here>"

b, err := job.GetOrCreateBoard(boardName, boardPrefix, template)
```

As you can see, it is the plugin's responsibility to assign the board a `name` and a `prefix`, which you will normally want to allow the user to specify through the config file. The `prefix` is added to the name of each flow in the board and helps to prevent collisions by allowing you to reuse the same board template to create multiple boards in a given account.

The board created by `Job.GetOrCreateBoard()` is automatically marked as read-only, so that users cannot manually modify it and cause your plugin to suddenly be faced with an unexpected layout. The creation of the board is handled intelligently, so that, if it already exists, it is not overwritten, and so that any differences between the layout that's on the API and the template is automatically managed for you. In other words, on a clean return from `GetOrCreateBoard()`, you are guaranteed to be given a board instance that conforms to the template every time.

### Creating tasks

Your next step will consist of creating one or more tasks. Tasks are very efficient, and you can have as many of them as you wish; it is therefore usually convenient to dedicate a single task to populating an individual flow to make things as simple as possible.

A task is simply a function that is called on a schedule. The function is synchronous and should return as soon as it has done its job. It will then be called again when the wait interval has passed. For example:

```go

func closure(job *job.Job) {
  // Do something here, then exit
}

// `r` is an instance of your plugin
r.AddTaskWithClosure(closure, 1.5 * time.Second)
```

It should be noted that the helper doesn't care where `closure` is defined as long as it can be executed. Therefore, nothing prevents you from passing a reference to a function a struct method—an easy way to give your closures a common context in which they can run.

**Important:** Remember that tasks run concurrently with each other; it is your responsibility to ensure that they do not enter into any race conditions (you can, however, generally consider every method of `Job` reentrant and thread-safe).

### Associating tasks with flows

Since most tasks will be responsible for populating flows, `PluginHelper` provides a set of methods that simplify this process. For example, suppose that you have a function called `fillValue()` that you want to send data every 5 seconds to the `value_x` flow of a board `b` that you have created:

```go
// `r` is an instance of your plugin
// On output, `err` will contain an error if the flow cannot be found
err := r.AddTaskWithClosureFromBoardForFlowWithTag(fillValue, time.Second * 5, b, "value_x")

func fillValue(job *job.Job, f *gotelemetry.Flow) {
  // Do something with the flow, then exit.
}
```

Note that `AddTaskWithClosureFromBoardForFlowWithTag()` will automatically handle board prefixes for you; therefore, you can safely reference flows by the tag they were given when a board template was created.

### Sending data to a flow

The Agent does not expose any Telemetry credentials directly to a plugin. Instead, you use the job associated with your plugin instance to send flow updates:

```go
func fillValue(job *job.Job, f *gotelemetry.Flow) {
  // Set some value into f

  data, err := f.ValueData()

  if !err {
    // Report an error
    return
  }

  data.Value = rand.Float64() * 10000

  job.PostFlowUpdate(f)
}
```

Internally, `PostFlowUpdate()` uses an asynchronous channel to coordinate all writes the Telemetry API. It will return immediately, but the data you submit may not be sent to the API until some time later, depending on the update interval set in the config file.

### Logging and reporting errors

Since plugins run asynchronously, `Job` also provides methods for reporting errors and for logging data:

```go
  if !err {
    // Report an error
    job.ReportError(gotelemetry.NewError(500, "Something has gone horribly wrong!"))
    return
  }

  job.Log("This is a log iem")
  job.Logf("This is also a log item with a %s", "parameter")
```

Note that `ReportError()` will not stop your plugin's execution.
