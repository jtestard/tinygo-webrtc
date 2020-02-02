/* eslint-env browser */
$(document).ready(() => {
  let pc = new RTCPeerConnection({
    iceServers: [
      {
        urls: 'stun:stun.l.google.com:19302'
      }
    ]
  })
  var log = msg => {
    $('#logs').append(msg + '<br>');
  }

  navigator.mediaDevices.getUserMedia({ video: true, audio: true })
      .then(stream => {
        stream.getTracks().forEach(track => pc.addTrack(track, stream));
        pc.createOffer().then(d => pc.setLocalDescription(d)).catch(log)
      }).catch(log)

  pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)
  pc.onicecandidate = event => {
    if (event.candidate === null) {
      $('#localSessionDescription').val(btoa(JSON.stringify(pc.localDescription)))
    }
  }
  let el;
  pc.ontrack = function (event) {
    el = document.createElement(event.track.kind);
    el.srcObject = event.streams[0];
    el.autoplay = true;
    el.controls = true;
    $('#remoteVideos').append(el);
  };

  window.startSession = () => {
    let sd = $('#remoteSessionDescription').val();
    if (sd === '') {
      return alert('Session Description must not be empty')
    }
    try {
      pc.setRemoteDescription(new RTCSessionDescription(JSON.parse(atob(sd))))
    } catch (e) {
      alert(e)
    }
  }

  window.closeSession = () => {
    success = () => {
      $('#remoteSessionDescription').val("");
    }
    fail = (err) => {
      alert(err.responseText)
    }
    pc.close();
    el.srcObject.getTracks().forEach(function(track) {
      track.stop();
    });
    el.remove();
    stream.getTracks().forEach(function(track) {
      track.stop();
    });
    $.post("/webrtc/close").done(success).fail(fail)
  }

  window.sendSession = () => {
    let sessionData = $('#localSessionDescription').val();
    success = (data) => {
      $('#remoteSessionDescription').val(data);
    }
    fail = (err) => {
      alert(err.responseText)
    }
    $.post("/webrtc/open", sessionData).done(success).fail(fail);
  }
})