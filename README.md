# record.go

Recorder frontend using [gpu-screen-recorder](https://git.dec05eba.com/gpu-screen-recorder).

## Examples

Only one instance of record.go is allowed at any time, which simplifies the CLI.

Record your screen, save to a timestamped file in ~/Videos:
```
record -d ~/Videos screen
```

End the previous recording with either ctrl+C or:
```
record
```

Toggle pause/unpause of a currently running session with:
```
record toggle
```

Use clipper/replay mode, like shadowplay:
```
record -d ~/Videos screen clipper
```
Optionally put a number after clipper for how many seconds you want the buffer length to be.

Saving clips is done the same way as pausing:
```
record toggle
# or
record clip
```
Ending the recording will not save a clip.

## Options

Run `record -h` for the list of options and recording targets.

## Todo

Some stuff, like setting video resolution to region resolution in region mode.
