 package main

 import (
	 "encoding/base64"
	 "encoding/json"
	 "fmt"
	 "github.com/pion/webrtc"
	 "github.com/pion/webrtc/pkg/media"
	 "github.com/pion/webrtc/pkg/media/ivfreader"
	 "html/template"
	 "io"
	 "io/ioutil"
	 "log"
	 "math/rand"
	 "net/http"
	 "os"
	 "sync"
	 "time"
 )

func main() {
	// Block forever
	http.HandleFunc("/", getWeb)
	http.HandleFunc("/webrtc/open", startWebRTCSession)
	http.HandleFunc("/webrtc/close", closeWebRTCSession)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	fmt.Println("now serving on localhost:8000")
	checkNoError(http.ListenAndServe(":8000", nil))
}

func checkNoError(err error) {
	if err != nil {
		panic(err)
	}
}

// GetWeb returns mirrorweb frontend
func getWeb(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("demo.html")
	if err != nil {
		log.Fatal(err)
	}

	tmpl.Execute(w, nil)
}

var peerConnection *webrtc.PeerConnection
var lock sync.Mutex

func closeWebRTCSession(w http.ResponseWriter, r *http.Request) {
	if peerConnection == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("session already closed/never opened"))
		return
	}
	checkNoError(peerConnection.Close())
	peerConnection = nil
}


func startWebRTCSession(w http.ResponseWriter, r *http.Request) {
	if peerConnection != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("session already started. Please close before re-opening"))
		return
	}

	// Everything below is the Pion WebRTC API! Thanks for using it ❤️.
	buf, err := ioutil.ReadAll(r.Body)
	checkNoError(err)

	// The mirrorweb rtc offer is sent over in the body of the request
	offer := webrtc.SessionDescription{}
	decode(string(buf), &offer)

	// We make our own mediaEngine so we can place the sender's codecs in it. Since we are echoing their RTP packet
	// back to them we are actually codec agnostic - we can accept all their codecs. This also ensures that we use the
	// dynamic media type from the sender in our answer.
	mediaEngine := webrtc.MediaEngine{}

	// Add codecs to the mediaEngine. Note that even though we are only going to echo back the sender's video we also
	// add audio codecs. This is because createAnswer will create an audioTransceiver and associated SDP and we currently
	// cannot tell it not to. The audio SDP must match the sender's codecs too...
	err = mediaEngine.PopulateFromSDP(offer)
	checkNoError(err)

	// Search for VP8 Payload type. If the offer doesn't support VP8 exit since
	// since they won't be able to decode anything we send them
	var payloadType uint8
	for _, videoCodec := range mediaEngine.GetCodecsByKind(webrtc.RTPCodecTypeVideo) {
		if videoCodec.Name == "VP8" {
			payloadType = videoCodec.PayloadType
			break
		}
	}
	if payloadType == 0 {
		panic("Remote peer does not support VP8")
	}

	// Create a new RTCPeerConnection
	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))
	peerConnection, err = api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	checkNoError(err)

	// Create a video track
	videoTrack, err := peerConnection.NewTrack(payloadType, rand.Uint32(), "video", "pion")
	checkNoError(err)
	_, err = peerConnection.AddTrack(videoTrack)
	checkNoError(err)

	go func() {
		// Open a IVF file and start reading using our IVFReader
		file, ivfErr := os.Open("output.ivf")
		checkNoError(ivfErr)

		ivf, header, ivfErr := ivfreader.NewWith(file)
		checkNoError(ivfErr)

		// Send our video file frame at a time. Pace our sending so we send it at the same speed it should be played back as.
		// This isn't required since the video is timestamped, but we will such much higher loss if we send all at once.
		sleepTime := time.Millisecond * time.Duration((float32(header.TimebaseNumerator)/float32(header.TimebaseDenominator))*1000)
		for {
			frame, _, ivfErr := ivf.ParseNextFrame()
			if ivfErr == io.EOF {
				break
			}
			checkNoError(ivfErr)

			time.Sleep(sleepTime)
			ivfErr = videoTrack.WriteSample(media.Sample{Data: frame, Samples: 90000})
			checkNoError(ivfErr)
		}
	}()

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	checkNoError(err)

	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	checkNoError(err)

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	checkNoError(err)

	// return the answer to the browser in base64
	_, err = w.Write([]byte(encode(answer)))
	checkNoError(err)
	fmt.Println("response sent to browser")

	// print the answer to the logs
	fmt.Println(encode(answer))
}

// Encode encodes the input in base64
// It can optionally zip the input before encoding
func encode(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(b)
}

// Decode decodes the input from base64
// It can optionally unzip the input after decoding
func decode(in string, obj interface{}) {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(b, obj)
	if err != nil {
		panic(err)
	}
}
