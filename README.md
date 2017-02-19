# flightcontrol

[![GoDoc](https://godoc.org/github.com/chriscasola/flightcontrol?status.svg)](https://godoc.org/github.com/chriscasola/flightcontrol)

Flightcontrol is a go package for managing a job queue. To use it you create a dispatcher, call Start, and
then AddJob for each task you want the dispatcher to handle. The number of workers created by the dispatcher
is configurable as is the max number of outstanding jobs.

See [the documentation](https://godoc.org/github.com/chriscasola/flightcontrol) on godoc.