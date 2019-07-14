# Simulation of fire mages DPS in WoW Vanilla

All values are from Light's Hope Northdale server database; tested and verified on custom private server

## Requirements

Simulation is written in golang - to run it you will need to install any version of golang (tested with `go v1.8.3`)

https://golang.org/dl/

## Running

There are 3 parameters inside `main.go` that decide how long in terms of real-time and simulation count it will be:

```
TimeScale  // x times to speed up simulation; more than 50x might be inaccurate
SimulationDuration  // How long should 1 simulation be (in realtime seconds; scaled by previous value); anything below 20 might be inaccurate
SimulationRunCount  // How many simulations to run in a row
```

All stats and strategies can be customized inside the code.. Addional logging for debug purposes can be enabled by making `LOG_*` constants `true`.

To run one instance:

`$ go run main.go`

Because simulations take a while to complete, its recommended to run multiple at a time, while saving the results in files:

`$ for i in {1..5}; do go run main.go > output/out$i.txt & done`

Its best to experiment with number of concurrent processes, but anything above 10 might make computer unusable.

## Notes

There are things missing like trinkets and spell penetration (it is set to have 0 resistances now)...
