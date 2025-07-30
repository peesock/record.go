package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"unicode"
)

var programName = "record"

type logger struct {
	status int
	verbosity int
}

func (l logger) log(format string, args ...any) {
	var newargs []any = make([]any, 1);
	newargs[0] = programName
	newargs = append(newargs, args...)
	fmt.Printf("%s: " + format + "\n", newargs...)
}

func (l logger) info(format string, args ...any) {
	if 1 >= l.status {
		l.log("[info] " + format, args...)
	}
}
func (l logger) warn(format string, args ...any) {
	if 2 >= l.status {
		l.log("[warn] " + format, args...)
	}
}
func (l logger) error(format string, args ...any) {
	if 3 >= l.status {
		l.log("[error] " + format, args...)
	}
}

var stateFile string

func main(){
	log := logger {
		status: 1,
	}
	var b bool
	stateFile, b = os.LookupEnv("XDG_RUNTIME_DIR")
	if !b {
		stateFile = "/run/user/" + strconv.Itoa(os.Getuid())
	}
	
	stat, err := os.Stat(stateFile);
	if err != nil {
		log.error("%v", err)
		os.Exit(1)
	}
	if !stat.IsDir() {
		log.error("'%s' not a directory.", stateFile)
		os.Exit(1)
	}

	stateFile = stateFile + "/" + programName

	fp, err := os.OpenFile(stateFile, os.O_CREATE, 0755);
	buf := make([]byte, 4096)
	n, err := fp.Read(buf)
	pid, err := strconv.ParseUint(string(buf[:n]), 10, 64)
	var begin bool
	if err != nil {
		begin = true
	} else {
		begin = false
	}

	log.info("%d, %v", pid, begin)

	if begin {
		// options
		var cmd *exec.Cmd;
		ch := make(chan []byte)
		go func(){
			cmd = exec.Command("sh", "-c", `xrandr --current | grep '\*'`)
			screenStr, err := cmd.Output()
			if err != nil {
				log.error("xrandr fail")
				os.Exit(1)
			}
			ch <- screenStr
		}()
		config := make(map[string] string)
		config["-c"] = "matroska"
		config["-q"] = "ultra"
		config["-k"] = "auto"

		var audiodevs []string
		audiodevs = append(audiodevs, "default_output")
		audiodevSet := false
		config["-ac"] = "opus"
		config["-ab"] = "256k"
		config["-v"] = "no"
		arglen := len(os.Args)
		if arglen < 2 {
			os.Exit(1)
		}
		var arg string
		n = 1
		for n < arglen {
			arg = os.Args[n]
			switch arg {
			case "-d":
				config["-ro"] = os.Args[n+1]
				n++
			case "-o":
				config["-o"] = os.Args[n+1]
				n++
			case "-a":
				if !audiodevSet {
					audiodevs = audiodevs[0:0]
					audiodevSet = true
				}
				audiodevs = append(audiodevs, os.Args[n+1])
				n++
			case "-ab":
				config["-ab"] = os.Args[n+1]
				n++
			case "-ac":
				config["-ac"] = os.Args[n+1]
				n++
			case "-an":
				audiodevs = audiodevs[0:0]
				delete(config, "-ab")
				delete(config, "-ac")
			case "-f":
				config["-f"] = os.Args[n+1]
				n++
			case "-s":
				config["-s"] = os.Args[n+1]
				n++
			case "-vq":
				config["-ab"] = os.Args[n+1]
				n++
			case "-vc":
				config["-ab"] = os.Args[n+1]
				n++
			case "-h", "-help", "--help":
				fmt.Println(
`opts:
-d dir -- set output dir
-o file -- set output file
-a device -- set audio device
-ab bitrate -- set audio bitrate
-ac codec -- set audio codec
-an -- don't use audio
-f rate -- set framerate
-s WxH -- set resolution
-vc codec -- set video codec
-vq quality -- set video quality
targets:
screen -- record first monitor
follow -- record followed windows
region -- select a region to record
portal -- use wayland screen recording portal
target flags:
clipper [seconds] -- shadowplay mode, default length = 60s`)
				return
			default:
				goto end
			}
			n++
		}
		end:

		if n == arglen {
			log.error("arg issue")
			os.Exit(1)
		}
		switch os.Args[n] {
		case "screen":
			config["-w"] = "screen"
		case "follow":
			config["-w"] = "focused"
		case "portal":
			config["-w"] = "portal"
		case "region":
			config["-w"] = "region"
			_, b = os.LookupEnv("WAYLAND_DISPLAY")
			var cmd *exec.Cmd
			if b {
				log.error("i don't support wayland yet....")
				os.Exit(1)
			} else {
				_, b = os.LookupEnv("DISPLAY")
				if b {
					cmd = exec.Command("hacksaw")
				} else {
					log.error("epic fail")
					os.Exit(1)
				}
			}
			region, err := cmd.Output()
			if err != nil {
				log.warn("Cancelled selection.")
				return
			}
			config["-region"] = string(region)
		}
		log.info("%v", config)

		if n+1 < arglen {
			switch os.Args[n+1] {
			case "clipper":
				if n+2 < arglen {
					secs := os.Args[n+2]
					secn, err := strconv.Atoi(secs)
					if secn < 5 || secn > 1200 || err != nil {
						log.error("bad args for clipper")
						os.Exit(1)
					}
					config["-r"] = os.Args[n+2]

					_, b := config["-o"]
					if !b {
						tmp, b := config["-ro"]
						if b {
							config["-o"] = tmp
						}
					}
				} else {
					config["-r"] = "60"
				}
			}
		}

		recordArgs := make([]string, 0, 8)
		for k, v := range config {
			recordArgs = append(recordArgs, k, v)
		}
		for _, v := range audiodevs {
			recordArgs = append(recordArgs, "-a", v)
		}
		xrandr := <- ch
		// vid framerate
		_, b := config["-f"]
		if !b {
			start := -1
			end := -1
			for i, v := range xrandr {
				if start < 0 {
					if v >= '0' && v <= '9' {
						start = i
					}
				} else {
					if v == '*' {
						end = i
						break
					}
					if unicode.IsSpace(rune(v)) {
						start = -1
					}
				}
			}
			framerate, _ := strconv.ParseFloat(string(xrandr[start:end]), 32)
			recordArgs = append(recordArgs, "-f", strconv.Itoa(int(math.Ceil(framerate))))
		}

		// vid resolution
		_, b = config["-s"]
		if !b {
			start := -1
			end := -1
			for i, v := range xrandr {
				if start < 0 {
					if !unicode.IsSpace(rune(v)) {
						start = i
					}
				} else {
					end = i
					if unicode.IsSpace(rune(v)) {
						break
					}
				}
			}
			recordArgs = append(recordArgs, "-s", string(xrandr[start:end]))
		} else {
			recordArgs = append(recordArgs, "-s", "0x0")
		}

		log.info("output: %v", recordArgs)
		cmd = exec.Command("gpu-screen-recorder", recordArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else {
	}
}
