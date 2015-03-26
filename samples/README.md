This directory contains a few samples to get you started with the Telemetry Agent.

In the jobs/ directory, you will find a minimalistic configuration file designed to run a job that updates a value flow once every second. Identically functional samples are provided for PHP, Ruby, and Node.js.

In the plugins/ directory, you will find two sample plugins; one (`random`) simply populates a value flow with a random value—think of this as a simple starting point for your own plugin. The other (`intercom`) offers up a more complex implementation that provides a wide range of functionality related to the public API offered by [Intercom](https://intercom.io).

For more information on writing your own jobs and plugins, you can consult the [Agent documentation](http://telemetry.readme.io/v1.0/docs/agents).