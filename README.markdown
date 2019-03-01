# blink1-concourse

Use a Blink1 device to indicate whether there are broken builds.

For a given Concourse team (TODO pipeline or job), get the status of the latest build of each job (grouped by pipeline). For each failed job, let the Blink1 device blink red once (TODO green if it is good, blue if pending etc).
