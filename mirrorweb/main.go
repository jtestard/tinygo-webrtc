 package main

 import (
	 "encoding/base64"
	 "encoding/json"
	 "fmt"
	 "github.com/pion/rtcp"
	 "github.com/pion/webrtc"
	 "html/template"
	 "io"
	 "io/ioutil"
	 "log"
	 "math/rand"
	 "net/http"
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

	videoCodecs := mediaEngine.GetCodecsByKind(webrtc.RTPCodecTypeVideo)
	if len(videoCodecs) == 0 {
		panic("Offer contained no video codecs")
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))

	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	// Create a new RTCPeerConnection
	peerConnection, err = api.NewPeerConnection(config)
	checkNoError(err)

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	checkNoError(err)

	// Create Track that we send video back to browser on
	outputTrack, err := peerConnection.NewTrack(videoCodecs[0].PayloadType, rand.Uint32(), "video", "pion")
	checkNoError(err)

	// Add this newly created track to the PeerConnection
	_, err = peerConnection.AddTrack(outputTrack)
	checkNoError(err)

	// Set a handler for when a new remote track starts, this handler copies inbound RTP packets,
	// replaces the SSRC and sends them back
	peerConnection.OnTrack(func(track *webrtc.Track, receiver *webrtc.RTPReceiver) {
		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		// This is a temporary fix until we implement incoming RTCP events, then we would push a PLI only when a viewer requests it
		doneChan := make(chan struct{})
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			loop:
			for {
				select {
					case <- ticker.C:
						errSend := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: track.SSRC()}})
						if errSend != nil {
							fmt.Println(errSend)
						}
					case <-doneChan:
						break loop
				}
			}
		}()

		fmt.Printf("Track has started, of type %d: %s \n", track.PayloadType(), track.Codec().Name)
		for {
			// Read RTP packets being sent to Pion
			rtp, readErr := track.ReadRTP()
			if readErr == io.EOF {
				doneChan <- struct{}{}
				break
			}
			checkNoError(readErr)

			// Replace the SSRC with the SSRC of the outbound track.
			// The only change we are making replacing the SSRC, the RTP packets are unchanged otherwise
			rtp.SSRC = outputTrack.SSRC()

			writeErr := outputTrack.WriteRTP(rtp)
			checkNoError(writeErr)
		}
	})
	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

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
