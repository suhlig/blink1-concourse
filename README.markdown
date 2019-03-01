# blink1-concourse

Indicate broken builds in a [Concourse](https://concourse-ci.org/) pipeline using a [Blink1](https://blink1.thingm.com/) device.

For a given Concourse target (TODO pipeline or job), get the status of the latest build of each job (grouped by pipeline). Then, for each failed job, let the Blink1 device blink red once (TODO green if it is good, blue if pending etc).

# Develop

1. Clone the project
1. Run `scripts/setup` to install dependencies
1. Run `scripts/build` in order to build the binary. Find it in `bin`.

For continuous development, use `scripts/iterate`.
