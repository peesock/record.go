# record.go

Recorder frontend using [gpu-screen-recorder](https://git.dec05eba.com/gpu-screen-recorder).

## Usage

Only one instance of record.go is allowed at any time, which simplifies the CLI.

To start recording:
```
record [-d,-o PATH] [OPTIONS] TARGET [ARGS]
```
Where:
- `-d` specifies output directory
- `-o` specifies output file
- `TARGET` is the recording type, "screen" for example
- `ARGS` which apply to the target.

To do an "action", currently either toggle pause or clip a video, run record with *any* arguments:
```
record hello
```
This will produce an error message if a video is not currently running. To suppress it, explicity
run:
```
record action
```

To end a recording, run record with *no* arguments:
```
record
```

## Examples

Record your screen, save to a timestamped file in ~/Videos:
```
record -d ~/Videos screen
```

You can use the same command to start and act on recordings:
```
record -d ~/Videos screen # start
sleep 5
record -d ~/Videos screen # pause
sleep 5
record -d ~/Videos screen # resume
sleep 5
record # kill
```

Use clipper/replay mode, like shadowplay:
```
record -d ~/Videos screen clipper
```
Optionally put a number after clipper for how many seconds you want the buffer length to be.

Saving clips is done the same way as pausing:
```
record args
```
*Ending the recording will not save a clip.*

## Options

Run `record -h` for the list of options and recording targets.

## Todo

Some stuff, like ffmpeging finished videos, user-supplied file namers, homecooking region mode,
apparently.
