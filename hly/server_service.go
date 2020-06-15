package hly

import (
	"fmt"
	"github.com/gorilla/websocket"
	vad "github.com/henryleu/vads/vad"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
	"vads/hly/util"
)

var upgrader = websocket.Upgrader{} // use default options

const debug = true

const requestTimeout = time.Second * 30
const chunkTimeout = time.Second * 5
const minFrameSize = 160 // 8000/1000*2*10
const sampleRate = 8000
const bytesPerSample = 2
const frameDuration = 20 // 20ms 10 20 30
const frameLen = frameDuration * 16

// mock to load config from file on boot
func getConfig() *vad.Config {
	c := vad.NewDefaultConfig()
	c.SilenceTimeout = 800   // 800 is the best value, test it before changing
	c.SpeechTimeout = 800    // 800 is the best value, test it before changing
	c.NoinputTimeout = 60000 // nearly ignore noinput case
	c.RecognitionTimeout = 20000
	c.VADLevel = 3     // 3 is the best value, test it before changing
	c.Multiple = false // recognition mode

	err := c.Validate()
	if err != nil {
		log.Fatalf("Config.Validate() error = %v", err)
	}
	return c
}

func sendCloseMessage(c *websocket.Conn, code int, msg string) {
	if code != websocket.CloseNormalClosure {
		c.SetWriteDeadline(time.Now().Add(writeWait))
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, msg))
	}
	time.Sleep(closeGracePeriod)
}

// return success response-message
func sendResponseMessage(wire *Wire, req *Request, status int, recognition *Recognition) {
	log.Print(recognition)
	res := req.NewSuccessResponse(status, recognition)
	err := wire.Send(res.Message())
	if err != nil {
		log.Printf(fmt.Sprintf("Wire.Send(responseMsg), error = %v", err))
		// when error on wire, ws connection cannot be closed gracefully any more
		return
	}
	wire.SendCloseMessage(websocket.CloseUnsupportedData, "")
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
			errMsg = fmt.Sprintf("fail to get request msg, error = %v\n", err)
		}
	case err = <-wire.ErrCh:
		errMsg = fmt.Sprintf("fail to get request msg, error = %v\n", err)
	case <-time.After(requestTimeout):
		errMsg = "fail to get request msg, error = timeout\n"
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

	chunkNo := 0
	frameNum := 0
loop_chunk:
	for {
		select {
		case msg := <-wire.MsgCh:
			chunk, err := msg.Chunk()
			if err != nil {
				errMsg = fmt.Sprintf("fail to get chunk msg, error = %v\n", err)
				break loop_chunk
			}
			err = chunk.DecodeAudio()
			if err != nil {
				errMsg = fmt.Sprintf("fail to decode chunk audio, error = %v\n", err)
				break loop_chunk
			}
			if req.CID != chunk.CID {
				errMsg = fmt.Sprintf("fail to decode chunk audio, error = %v\n", err)
				break loop_chunk
			}
			cno, err := strconv.Atoi(chunk.NO)
			if err != nil {
				errMsg = fmt.Sprintf("fail to unmarshal chunk no (%v), error = %v\n", chunk.NO, err)
				break loop_chunk
			}
			chunkNo++
			if chunkNo != cno {
				errMsg = fmt.Sprintf("fail to validate chunk no, want %d, got %d\n", chunkNo, cno)
				break loop_chunk
			}
			chunkSize := len(chunk.Data)
			if chunkSize%minFrameSize != 0 {
				errMsg = fmt.Sprintf("fail to validate chunk data, the size is %d\n", chunkSize)
				break loop_chunk
			}
			log.Printf("chunk no %v\n", chunk.NO)

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
				if !detector.Working() {
					log.Printf("detector is stopped for session [%v] after %v chunks\n", chunk.CID, chunk.NO)
					break loop_chunk
				}
				if err != nil {
					errMsg = fmt.Sprintf("fail to process frame in chunk NO[%v] of session[%v], error = %v\n", chunk.NO, chunk.CID, err)
					break loop_chunk
				}
				frameNum++
			} // end loop frame
			// go on looping more chunks
		case err = <-wire.ErrCh:
			errMsg = fmt.Sprintf("fail to get chunk msg, error = %v\n", err)
			break loop_chunk
		case <-time.After(chunkTimeout):
			errMsg = "fail to get chunk msg, error = timeout\n"
			break loop_chunk
		}
	} // end loop chunk

	// check error
	if errMsg != "" {
		sendErrorResponse(wire, req, errMsg)
		return
	}

	// new a clip file name here
	var voicePath string = "/mnt/test-%v-%v.wav"
	log.Printf("frame number %v\n", frameNum)
events_loop:
	for e := range detector.Events {
		switch e.Type {
		case vad.EventVoiceBegin:
			// ignore handling
		case vad.EventVoiceEnd:
			//f, err := ioutil.TempFile("", fmt.Sprintf("clip-%v-*.wav", req.CID))
			t := time.Now()
			voicePath = fmt.Sprintf(voicePath, req.CID, t.Format("20060102150405"))
			f, err := os.Create(voicePath)
			if err != nil {
				errMsg = fmt.Sprintf("fail to save clip, fs.Open() error = %v\n", err)
				log.Print(errMsg)
				break events_loop
			}
			e.Clip.SaveToWriter(f)
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
	log.Printf("voice_path: %s\n", voicePath)
	asrText := util.AsrClient(voicePath)
	postData := map[string]interface{}{
		"user_id":  req.CID,
		"robot_id": "4a44a2992fbaf64d5c19fb1b192f45c8",
		"input":    asrText,
		"token":    "21c7d084b200a17c9641c83d4697fde9",
	}
	if flowReturn, err := util.FlowUtilSay(postData); err == nil {
		flowData := flowReturn.(map[string]interface{})
		recog := Recognition{
			AnswerText: flowData["slot_output"].(string),
			AudioText:  asrText,
			AudioNum:   strings.Replace(flowData["output_command"].(string), "\r\n", "", -1),
		}
		var status int
		if flowData["flow_end"] == true {
			status = 1
		} else {
			status = 0
		}
		msg := req.NewSuccessResponse(status, &recog)
		log.Printf("msg.Message: %s\n", msg.Message())
		err = wire.Send(msg.Message())
		if err != nil {
			log.Fatalf("Wire.Send(requestMsg) error = %v", err)
		}
		log.Println(recog)
	}
	sendCloseMessage(c, websocket.CloseNormalClosure, "")
}

// Home is the handler for testing in webrtc client
func Home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/mrcp")
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };

});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server,
"Send" to send a message to the server and "Close" to close the connection.
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
