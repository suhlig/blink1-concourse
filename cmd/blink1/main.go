package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/concourse/atc"
	"github.com/concourse/fly/rc"
	"github.com/concourse/go-concourse/concourse"
	"github.com/dgrijalva/jwt-go"
	blink1 "github.com/hink/go-blink1"
	flags "github.com/jessevdk/go-flags"
)

func main() {
	// TODO like "fly build":
	// --pipeline PIPELINE (optional) just this pipeline; fail if it doesn't exist
	// --job PIPELINE/JOB (optional) just this job; fail if it doesn't exist
	// --team (optional) just the currently targeted team

	var options struct {
		Target string `short:"t" long:"target" description:"Concourse target name" required:"true"`
	}

	_, err := flags.NewParser(&options, flags.Default).Parse()

	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	target, err := rc.LoadTarget(rc.TargetName(options.Target), false)

	if err != nil {
		die(err)
	}

	err = target.Validate()

	if err != nil {
		die(err)
	}

	err = validateToken(target.Token())

	if err != nil {
		fmt.Printf("Warning: %s.\n", err.Error())
	}

	pipelines, err := target.Team().ListPipelines()

	if err != nil {
		die(err)
	}

	var states []blink1.State

	if len(pipelines) == 0 {
		fmt.Println("Warning: no pipelines found")
		states = append(states,
			blink1.State{Red: 96, Green: 64, Blue: 0, Duration: time.Duration(100) * time.Millisecond},
			blink1.State{Duration: time.Duration(100) * time.Millisecond},
			blink1.State{Red: 96, Green: 64, Blue: 0, Duration: time.Duration(100) * time.Millisecond},
			blink1.State{Duration: time.Duration(100) * time.Millisecond},
			blink1.State{Red: 96, Green: 64, Blue: 0, Duration: time.Duration(100) * time.Millisecond},
		)
	}

	states = append(states, getPipelineStates(target.Team(), pipelines)...)

	err = render(states)

	if err != nil {
		die(err)
	}
}

func validateToken(tToken *rc.TargetToken) error {
	if tToken == nil || tToken.Value == "" {
		return errors.New("Not logged in")
	}

	if tToken != nil {
		_, err := jwt.Parse(tToken.Value, func(token *jwt.Token) (interface{}, error) {
			return nil, token.Claims.Valid()
		})

		if err != nil && err.Error() != jwt.ErrInvalidKeyType.Error() {
			return err
		}
	}

	return nil
}

func getPipelineStates(team concourse.Team, pipelines []atc.Pipeline) []blink1.State {
	var states []blink1.State

	for _, p := range pipelines {
		jobs, err := team.ListJobs(p.Name)
		if err != nil {
			fmt.Printf("  %e\n", err)
		}

		states = append(states, getJobStates(jobs)...)
	}

	return states
}

func getJobStates(jobs []atc.Job) []blink1.State {
	var states []blink1.State

	for _, j := range jobs {
		if j.Paused { // blue
			fmt.Printf("  %s: paused\n", j.Name)
			states = append(states, blink1.State{Red: 0, Green: 0, Blue: 128, Duration: time.Duration(50) * time.Millisecond})
			continue
		}

		if j.FinishedBuild == nil {
			fmt.Printf("  %s: no finished builds\n", j.Name)
			continue
		}

		status := j.FinishedBuild.Status

		switch status {
		case "started": // grey striped
			states = append(states,
				blink1.State{Red: 64, Green: 64, Blue: 64, Duration: time.Duration(50) * time.Millisecond},
				blink1.State{Duration: time.Duration(20) * time.Millisecond})
		case "pending": // grey
			states = append(states, blink1.State{Red: 32, Green: 32, Blue: 32, Duration: time.Duration(50) * time.Millisecond})
		case "succeeded": // green
		// no news are good news
		case "failed": // red
			states = append(states, blink1.State{Red: 128, Green: 0, Blue: 0, Duration: time.Duration(50) * time.Millisecond})
		case "errored": // orange
			states = append(states, blink1.State{Red: 128, Green: 64, Blue: 0, Duration: time.Duration(50) * time.Millisecond})
		case "aborted": // brown
			states = append(states, blink1.State{Red: 139, Green: 87, Blue: 42, Duration: time.Duration(50) * time.Millisecond})
		default:
			fmt.Printf("  Error: Status '%s' has no color mapping\n", status)
		}

		if status != "succeeded" {
			fmt.Printf("%s/%s: %s since %s\n",
				j.PipelineName,
				j.FinishedBuild.JobName,
				j.FinishedBuild.Status,
				time.Unix(j.FinishedBuild.EndTime, 0).Format(time.RFC3339))

			// pause between patterns
			states = append(states, blink1.State{Duration: time.Duration(50) * time.Millisecond})
		}

		states = append(states)
	}

	return states
}

func render(states []blink1.State) error {
	device, err := blink1.OpenNextDevice()
	defer device.Close()

	if err != nil {
		return err
	}

	device.RunPattern(&blink1.Pattern{
		Repeat:      0,
		RepeatDelay: time.Duration(100) * time.Millisecond,
		States:      states,
	})

	return nil
}

func die(err error) {
	fmt.Printf("Error: %s", err.Error())
	os.Exit(1)
}
