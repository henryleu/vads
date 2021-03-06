package hly

import (
	"fmt"
	"path"

	"github.com/gorilla/websocket"
	"github.com/henryleu/go-vad"
	"github.com/henryleu/vads/hly/util"

	"log"
	"net/http"
	"os"
	"time"
)

var upgrader = websocket.Upgrader{} // use default options

const debug = true

const requestTimeout = time.Second * 30
const chunkTimeout = time.Second * 2
const minFrameSize = 160 // 8000/1000*2*10
const sampleRate = 8000
const bytesPerSample = 2
const frameDuration = 20 // 20ms 10 20 30
const frameLen = frameDuration * 16

// mock to load config from file on boot
func getConfig() *vad.Config {
	c := vad.NewDefaultConfig()
	c.SilenceTimeout = 400   // 800 is the best value, test it before changing
	c.SpeechTimeout = 400    // 800 is the best value, test it before changing
	c.NoinputTimeout = 20000 // nearly ignore noinput case
	c.RecognitionTimeout = 10000
	c.VADLevel = 2     // 3 is the best value, test it before changing
	c.Multiple = false // recognition mode
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	err := c.Validate()
	if err != nil {
		log.Fatalf("Config.Validate() error = %v", err)
	}
	return c
}

func getVoiceDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("fail to get wd, error: %v", err)
		return "", err
	}
	dir := path.Join(cwd, "tmp")
	_, err = os.Stat(dir)
	if err != nil {
		log.Printf("voice dir %q doesn't exist, error: %v", dir, err)
	}
	return dir, nil
}

func sendCloseMessage(c *websocket.Conn, code int, msg string) {
	if code != websocket.CloseNormalClosure {
		c.SetWriteDeadline(time.Now().Add(writeWait))
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, msg))
	}
	time.Sleep(closeGracePeriod)
}

func sendErrorResponse(wire *Wire, req *Request, errMsg string) {
	log.Print(errMsg)
	res := req.NewErrorResponse(errMsg)
	err := wire.Send(res.Message())
	if err != nil {
		log.Printf(fmt.Sprintf("Wire.Send(responseMsg), error = %v", err))
		// when error on wire, ws connection cannot be closed gracefully any more
		return
	}
	wire.SendCloseMessage(websocket.CloseUnsupportedData, errMsg)
}

// HandleMRCP is the handler for websocket
func HandleMRCP(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer c.Close()

	wire := NewWire(c)
	go wire.ServerReceive()

	log.Printf("frame length %v\n", frameLen)
	var req *Request
	var errMsg string
	select {
	case msg := <-wire.MsgCh:
		req, err = msg.Request()
		if err != nil {
			errMsg = fmt.Sprintf("fail 001 to get request msg, error = %v\n", err)
		}
	case err = <-wire.ErrCh:
		errMsg = fmt.Sprintf("fail to get request msg, error = %v\n", err)
	case <-time.After(requestTimeout):
		errMsg = "fail 002 to get request msg, error = timeout\n"
	}

	if errMsg != "" {
		log.Print(errMsg)
		wire.SendCloseMessage(websocket.CloseUnsupportedData, errMsg)
		return
	}

	config := getConfig()
	detector := config.NewDetector()
	detector.SampleRate = sampleRate
	detector.BytesPerSample = bytesPerSample
	detector.FrameDuration = frameDuration
	err = detector.Init()
	if err != nil {
		errMsg = fmt.Sprintf("Detector.Init() error = %v", err)
		log.Print(errMsg)
		wire.SendCloseMessage(websocket.CloseUnsupportedData, errMsg)
		return
	}
	log.Println("BytesPerFrame", detector.BytesPerFrame())
	chunkNo := 0
loop_chunk:
	for {
		select {
		case msg := <-wire.MsgCh:
			chunk, err := msg.Chunk()
			if err != nil {
				errMsg = fmt.Sprintf("fail  003 to get chunk msg, error = %v\n", err)
				break loop_chunk
			}
			err = chunk.DecodeAudio()
			if err != nil {
				errMsg = fmt.Sprintf("fail 004 to decode chunk audio, error = %v\n", err)
				break loop_chunk
			}
			if req.CID != chunk.CID {
				errMsg = fmt.Sprintf("fail 005 to decode chunk audio, error = %v\n", err)
				break loop_chunk
			}
			chunkNo++
			if chunkNo != chunk.NO {
				errMsg = fmt.Sprintf("fail to validate chunk no, want %d, got %d\n", chunkNo, chunk.NO)
				break loop_chunk
			}
			chunkSize := len(chunk.Data)
			if chunkSize%minFrameSize != 0 {
				errMsg = fmt.Sprintf("fail to validate chunk data, the size is %d\n", chunkSize)
				break loop_chunk
			}

			// process chunks
			data := chunk.Data
			frame := data // chunk data is a slice with 1280 bytes
			for {
				l := len(data)
				if l <= 0 {
					break
				}
				frame = data[:frameLen] // a slice with 320 bytes
				data = data[frameLen:]
				err := detector.Process(frame)
				if err != nil {
					errMsg = fmt.Sprintf("fail to process frame in chunk NO[%v] of session[%v], error = %v\n", chunk.NO, chunk.CID, err)
					detector.Finalize()
					break loop_chunk
				}
				if !detector.Working() {
					log.Printf("detector is stopped for session [%v] after %v chunks\n", chunk.CID, chunk.NO)
					detector.Finalize()
					break loop_chunk
				}
			} // end loop frame
			// go on looping more chunks
		case err = <-wire.ErrCh:
			errMsg = fmt.Sprintf("fail 006 to get chunk msg, error = %v\n", err)
			break loop_chunk
		case <-time.After(chunkTimeout):
			detector.Finalize()
			// errMsg = "fail 007 to get chunk msg, error = timeout\n"
			break loop_chunk
		}
	} // end loop chunk

	// check error
	if errMsg != "" {
		sendErrorResponse(wire, req, errMsg)
		return
	}

	// new a clip file name here
	voiceDir, _ := getVoiceDir()
	fileTpl := "hly-%v-%v.wav"
	voiceTpl := path.Join(voiceDir, fileTpl)
	voicePath := ""
	// var voicePath string = "/mnt/voice/hly-%v-%v.wav"

	//var infoMsg = make(map[string]interface{})
events_loop:
	for e := range detector.Events {
		switch e.Type {
		case vad.EventVoiceBegin:
			// 根据被叫获取当前流程信息以及场景信息
			//postData := map[string]interface{}{
			//	"mobile": req.Business.Called,
			//}
			//if flowInfo, err := util.FlowInfoByNumber(postData); err == nil {
			//	infoMsg = flowInfo.(map[string]interface{})
			//}
		case vad.EventVoiceEnd:
			//f, err := ioutil.TempFile("", fmt.Sprintf("clip-%v-*.wav", req.CID))
			t := time.Now()
			log.Println(chunkNo)

			// detected clip
			voicePath = fmt.Sprintf(voiceTpl, req.CID, t.Format("20060102150405001"))
			f, err := os.Create(voicePath)
			if err != nil {
				errMsg = fmt.Sprintf("fail to save clip, fs.Open() error = %v\n", err)
				log.Print(errMsg)
				break events_loop
			}
			detector.Clip.SaveToWriter(f)
			detector.Clip.PrintDetail()
			log.Println("detector.SpeechTimeout", detector.SpeechTimeout)
			log.Println("detector.SilenceTimeout", detector.SilenceTimeout)
			log.Println("detector.BytesPerFrame", detector.BytesPerFrame())
			log.Println("clip degest", detector.Clip.GenerateDigest())

			// total clip
			voicePath = fmt.Sprintf(voiceTpl, req.CID, t.Format("20060102150405002"))
			f, err = os.Create(voicePath)
			if err != nil {
				errMsg = fmt.Sprintf("fail to save clip, fs.Open() error = %v\n", err)
				log.Print(errMsg)
				break events_loop
			}
			tc := detector.GetTotalClip()
			tc.SaveToWriter(f)
			tc.PrintDetail()
			log.Println("total clip degest", tc.GenerateDigest())

			log.Printf("succeed to save clip %v for session %v\n", f.Name(), req.CID)
			break events_loop
		case vad.EventNoinput:
			errMsg = fmt.Sprintf("fail to detect noinput speech for session %v\n", req.CID)
			break events_loop
		default:
			log.Printf("illegal event type %v\n", e.Type)
		}
	}
	// check error
	if errMsg != "" {
		sendErrorResponse(wire, req, errMsg)
		return
	}
	// todo asr and nlp here
	//asrText := util.AsrClient(voicePath)
	//postData := map[string]interface{}{
	//	"user_id":   req.CID,
	//	"token":     infoMsg["flow_token"],
	//	"robot_id":  infoMsg["robot_id"],
	//	"parameter": infoMsg["parameter"],
	//	"input":     asrText,
	//}
	//if flowReturn, err := util.FlowUtilSay(postData); err == nil {
	//	flowData := flowReturn.(map[string]interface{})
	//	output_command := strings.Replace(flowData["output_command"].(string), "\r\n", "", -1)
	//	arr := strings.Split(output_command, ".")
	//	recog := Recognition{
	//		AnswerText: flowData["user_label"].(string),
	//		AudioText:  asrText,
	//		AudioNum:   arr[0],
	//	}
	//	var status int
	//	if flowData["flow_end"] == true {
	//		status = 1
	//	} else {
	//		status = 0
	//	}
	//	msg := req.NewSuccessResponse(status, &recog)
	//	log.Printf("msg.Message: %s\n", msg.Message())
	//	err = wire.Send(msg.Message())
	//	if err != nil {
	//		log.Fatalf("Wire.Send(requestMsg) error = %v", err)
	//	}
	//} else {
	//	sendErrorResponse(wire, req, errMsg)
	//	return
	//}
	//  todo 返回语音识别结果
	asrText := util.AsrClient(voicePath)
	recog := Recognition{
		AnswerText: "",
		AudioText:  asrText,
		AudioNum:   "",
	}
	msg := req.NewSuccessResponse(0, &recog)
	log.Printf("msg.Message: %s\n", msg.Message())
	err = wire.Send(msg.Message())
	if err != nil {
		log.Fatalf("Wire.Send(requestMsg) error = %v", err)
	}
	sendCloseMessage(c, websocket.CloseNormalClosure, "")
}
