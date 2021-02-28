package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"gitee.com/rocket049/play-all-mp3/mpg123"

	"github.com/hajimehoshi/oto"
)

func main() {
	var h = flag.Bool("h", false, "显示帮助信息。")
	var notContinue = flag.Bool("nc", false, "从头播放，不读取播放进度")
	var fileContinue = flag.Bool("fc", false, "从上次播放的文件头部开始播放")
	var from = flag.String("from", "", "从该文件开始播放")
	flag.Parse()

	if *h || flag.Arg(0) == "" {
		fmt.Print("用法：play-all-mp3 <保存着MP3文件的目录>\n")
		flag.Usage()
		return
	}

	// 捕获信号

	dirName := flag.Arg(0)
	info, err := os.Stat(dirName)
	chk(err)
	if info.IsDir() == false {
		fmt.Println("Please input dir name")
		return
	}
	var mp3List []string
	var skip int32
	if len(*from) > 0 {
		mp3List, skip, err = getListWithName(dirName, *from)
	} else {
		mp3List, skip, err = getListWithPos(dirName, *notContinue)

		fmt.Println("Playing.  Press Ctrl-C to stop.")
		//修正进度
		var fix int32 = 10
		if skip > fix {
			skip -= fix
		} else {
			skip = 0
		}
		if *fileContinue {
			skip = 0
		}
	}

	for i := range mp3List {
		if playFile2(filepath.Join(dirName, mp3List[i]), skip) == -1 {
			break
		}
		skip = 0
	}
}

type PlayRecord struct {
	File string
	Pos  int32
}

//使用 portaudio
/*
func playFile(fileName string, skip int32) int {
	fmt.Printf("Play File:%s\nPos:%d\n", fileName, skip)
	var Pos int32
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	// create mpg123 decoder instance
	decoder, err := mpg123.NewDecoder("")
	chk(err)
	// 打开文件
	chk(decoder.Open(fileName))
	defer decoder.Close()

	// get audio format information
	rate, channels, _ := decoder.GetFormat()

	// make sure output format does not change
	decoder.FormatNone()
	decoder.Format(rate, channels, mpg123.ENC_SIGNED_16)

	portaudio.Initialize()
	defer portaudio.Terminate()
	out := make([]int16, 8192)
	// 用前面创建的channels打开流
	stream, err := portaudio.OpenDefaultStream(0, channels, float64(rate), len(out), &out)
	chk(err)
	defer stream.Close()

	chk(stream.Start())
	defer stream.Stop()
	atomic.StoreInt32(&Pos, 0)
	for {
		audio := make([]byte, 2*len(out))
		// 从decoder读出数据到audio
		_, err = decoder.Read(audio)
		if err == mpg123.EOF {
			break
		}
		chk(err)
		if atomic.LoadInt32(&Pos) < skip {
			atomic.AddInt32(&Pos, 1)
			continue
		}

		// 转换字节序，输出到out，这一步后数据就已经从decoder传送到了stream里
		chk(binary.Read(bytes.NewBuffer(audio), binary.LittleEndian, out))
		chk(stream.Write())
		atomic.AddInt32(&Pos, 1)
		select {
		case <-sig:
			fmt.Printf("File:%s\nPos:%d\n", fileName, Pos)
			updateRecord(fileName, atomic.LoadInt32(&Pos))
			return -1
		default:
		}
	}
	updateRecord(fileName, -1)
	return 0
}*/

func updateRecord(fn string, pos int32) {
	r := PlayRecord{File: filepath.Base(fn), Pos: pos}
	data, err := json.Marshal(&r)
	chk(err)
	recordName := filepath.Join(filepath.Dir(fn), "musicplay.log")
	ioutil.WriteFile(recordName, data, 0644)
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

func arrayReduce(a []string) []string {
	names := make([]string, 0, len(a))
	for _, name := range a {
		if strings.HasSuffix(strings.ToLower(name), ".mp3") {
			names = append(names, name)
			//fmt.Println(name)
		}
	}
	return names
}

func getListWithName(dir, name string) (txtList []string, pos int32, err error) {
	d, err := os.Open(dir)
	if err != nil {
		return
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return
	}

	names = arrayReduce(names)

	sort.Strings(names)

	txtList = names
	pos = 0
	err = nil

	for i := range names {
		if names[i] == name {
			txtList = names[i:]
			break
		}
	}

	return
}

func getListWithPos(dir string, nc bool) (txtList []string, pos int32, err error) {
	d, err := os.Open(dir)
	if err != nil {
		return
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return
	}

	names = arrayReduce(names)

	sort.Strings(names)

	if nc {
		txtList = names
		pos = 0
		err = nil
		return
	}

	var playLog PlayRecord

	recordName := filepath.Join(dir, "musicplay.log")
	logData, err := ioutil.ReadFile(recordName)
	if err == nil {
		err = json.Unmarshal(logData, &playLog)
		if err != nil {
			return
		}
	}

	pos = playLog.Pos
	var n int
	for n = 0; n < len(names); n++ {
		if names[n] == playLog.File {
			break
		}
	}
	if pos < 0 {
		pos = 0
		n++
	}
	if n >= len(names) {
		n = 0
	}
	txtList = names[n:]
	return
}

var otoCtx *oto.Context
var otoOnce sync.Once

//使用 oto
func playFile2(fileName string, skip int32) int {
	fmt.Printf("Play File:%s\nPos:%d\n", fileName, skip)
	var Pos int32
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	// create mpg123 decoder instance
	decoder, err := mpg123.NewDecoder("")
	chk(err)
	// 打开文件
	chk(decoder.Open(fileName))
	defer decoder.Close()

	// get audio format information
	rate, channels, _ := decoder.GetFormat()

	// make sure output format does not change
	decoder.FormatNone()
	decoder.Format(rate, channels, mpg123.ENC_SIGNED_16)

	var bufLen = 8192 * 2
	// 用前面创建的channels打开流
	otoOnce.Do(func() {
		var err error
		otoCtx, err = oto.NewContext(int(rate), channels, 2, bufLen)
		chk(err)
	})

	player := otoCtx.NewPlayer()
	defer player.Close()

	atomic.StoreInt32(&Pos, 0)
	for {
		audio := make([]byte, bufLen)
		// 从decoder读出数据到audio
		n, err := decoder.Read(audio)
		if err == mpg123.EOF {
			break
		}
		chk(err)
		if atomic.LoadInt32(&Pos) < skip {
			atomic.AddInt32(&Pos, 1)
			continue
		}

		_, err = player.Write(audio[:n])
		chk(err)
		atomic.AddInt32(&Pos, 1)
		select {
		case <-sig:
			fmt.Printf("File:%s\nPos:%d\n", fileName, Pos)
			updateRecord(fileName, atomic.LoadInt32(&Pos))
			return -1
		default:
		}
	}
	updateRecord(fileName, -1)
	return 0
}
