# record.go

Recorder frontend using [gpu-screen-recorder](https://git.dec05eba.com/gpu-screen-recorder).

## Examples

Only one instance of record.go is allowed at any time, which simplifies the CLI.

Record your screen, save to a timestamped file in ~/Videos:
```
record -d ~/Videos screen
```

End the previous recording with ctrl+C or with:
```
record kill
```
However to make keybinds easier, you can use any argument list besides "kill" to do the same thing.
But if there's nothing to kill, it will be treated as if you want to start a new recording.

In other words you can use the same command to start and stop recordings:
```
record -d ~/Videos screen # start
sleep 5
record -d ~/Videos screen # stop
```

Toggle pause/unpause of a currently running session by using no arguments:
```
record
```

Use clipper/replay mode, like shadowplay:
```
record -d ~/Videos screen clipper
```
Optionally put a number after clipper for how many seconds you want the buffer length to be.

Saving clips is done the same way as pausing:
```
record
```
*Ending the recording will not save a clip.*

## Options

Run `record -h` for the list of options and recording targets.

## Todo

Some stuff, like ffmpeging finished videos, user-supplied file namers, quiet mode, homecooking
region mode, apparently.
