package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
	"unicode"
)

var programName = "record"

type logger struct {
	status int
	verbosity int
}
var log = logger {
	status: 1,
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

func notify(msg string){
	exec.Command("notify-send", "-t", "2000", programName, msg).Start()
}

func recordHook(path string){
	if outDir == "" {
		return // explicit output
	}
	t := time.Now()
	ext := "mkv"
	name := fmt.Sprintf("%d%02d%02d-%02d%02d%02d-%d",
	t.Year(), t.Month(), t.Day(),
	t.Hour(), t.Minute(), t.Second(),
	t.Nanosecond() / 1000000)
	err := os.Rename(path, filepath.Dir(path) + "/" + name + "." + ext)
	if err != nil {
		log.error("%v", err)
		os.Exit(1)
	}
}

var stateFile string
var stateFifo string
var config map[string] string
var outDir string

func main(){

	var b bool
	stateFile, b = os.LookupEnv("XDG_RUNTIME_DIR")
	if !b {
		stateFile = "/run/user/" + strconv.Itoa(os.Getuid())
	}

	_, err := os.Stat(stateFile);
	if err != nil {
		log.error("%v", err)
		os.Exit(1)
	}

	stateFifo = stateFile + "/" + programName + ".status"
	stateFile = stateFile + "/" + programName

	// determine status of fifo
	fp, err := os.OpenFile(stateFifo, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		if os.IsNotExist(err) {
			err = syscall.Mknod(stateFifo, syscall.S_IFIFO|0666, 0)
			if err != nil {
				log.error("%v", err)
			}
		} else {
			log.error("%v", err)
		}
	}
	n, _ := fp.Read([]byte{0})
	if n > 0 { // has a writer; there is a process
		buf, err := os.ReadFile(stateFile)
		var pid int
		if err != nil {
			if os.IsNotExist(err) {
				pid = 0
				goto next
			} else {
				log.error("%v", err)
			}
			os.Exit(1)
		}
		pid, _ = strconv.Atoi(string(buf))
		next:
		log.info("pid: %d", pid)
		if len(os.Args) == 3 {
			switch os.Args[2] {
			case "regular":
			case "replay":
			default:
				goto skip
			}
			// called with recorder.
			err = os.WriteFile(stateFile, []byte(os.Args[1]), 0640)
			if err != nil {
				log.error("recorder write error: %v", err)
			}
			log.info("Sending SIGUSR2")
			syscall.Kill(pid, syscall.SIGUSR2)
			return
		}
		skip:
		// if stopping, sigint
		// if pausing/playing or clipping, sigusr1
		if len(os.Args) > 1 {
			switch os.Args[1] {
			case "toggle", "clip":
				log.info("Sending SIGUSR1")
				syscall.Kill(pid, syscall.SIGUSR1)
			}
		} else {
			log.info("Sending SIGINT")
			syscall.Kill(pid, syscall.SIGINT)
		}
		return
	}

	// write to the fifo to show we are alive
	go func(){
		fp, _ := os.OpenFile(stateFifo, os.O_WRONLY, 0)
		buf := make([]byte, 1024)
		for {
			fp.Write(buf)
		}
	}()

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
	config = make(map[string] string)
	config["-sc"] = programName
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
			outDir = os.Args[n+1]
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

	// default output file name
	_, b = config["-o"]
	if !b {
		if outDir == "" {
			log.error("need an output dir or file.")
			os.Exit(1)
		}
		fp, err := os.CreateTemp(outDir, "vid-")
		if err != nil {
			log.error("couldn't create tmpfile: %v", err)
			os.Exit(1)
		}
		fp.Close()
		config["-o"] = fp.Name()
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
				if outDir == "" {
					log.error("must set output dir for clipper")
					os.Exit(1)
				}
				config["-o"] = outDir
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
	_, b = config["-f"]
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

	cmd = exec.Command("gpu-screen-recorder", recordArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	waiting := true
	// INT, TERM to finish recording or stop clipping
	// USR1 to pause recording or make a clip
	// USR2 is sent by recorder to say it finished (and wrote to statefile)
	go func(){
		paused := false
		for sig := range sigs {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				syscall.Kill(cmd.Process.Pid, syscall.SIGINT)
			case syscall.SIGUSR1:
				_, b = config["-r"]
				if b {
					// clip
					syscall.Kill(cmd.Process.Pid, syscall.SIGUSR1)
					notify("Clipped")
				} else {
					// toggle pause
					syscall.Kill(cmd.Process.Pid, syscall.SIGUSR2)
					paused = !paused
					if paused {
						notify("Paused")
					} else {
						notify("Resumed")
					}
				}
			case syscall.SIGUSR2:
				waiting = false
				// finish up...
				path, err := os.ReadFile(stateFile)
				if err != nil {
					log.error("%v", err)
				}
				recordHook(string(path))
				done <- true
			}
		}
	}()

	cmd.Start()
	go func(){
		cmd.Wait()
		if waiting {
			time.Sleep(time.Millisecond * 1000)
			os.Exit(cmd.ProcessState.ExitCode())
		}
	}()

	err = os.WriteFile(stateFile, []byte(strconv.Itoa(os.Getpid())), 0640)
	if err != nil {
		log.error("pid file writing error: %v", err)
		cmd.Process.Kill()
		os.Exit(1)
	}

	<- done

	err = os.WriteFile(stateFile, []byte{}, 0640)
	if err != nil {
		log.error("%v", err)
		os.Exit(1)
	}
}
