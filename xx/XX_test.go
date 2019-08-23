package xx

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
	proto "github.com/gogo/protobuf/proto"
	pb "github.com/ChainSafe/go-libp2p-noise/pb"
	"github.com/libp2p/go-libp2p-core/crypto"
)

func TestGetHkdf(t *testing.T) {
	ck := [32]byte{}
	ckBytes, err := hex.DecodeString("4e6f6973655f58585f32353531395f58436861436861506f6c795f53484132353600000000000000000000000000000000000000000000000000000000000000")	
	if err != nil {
		t.Fatal(err)
	}
	copy(ck[:], ckBytes)

	ikm, err := hex.DecodeString("a3eae50ea37a47e8a7aa0c7cd8e16528670536dcd538cebfd724fb68ce44f1910ad898860666227d4e8dd50d22a9a64d1c0a6f47ace092510161e9e442953da3")
	if err != nil {
		t.Fatal(err)
	}

	a, b, c := getHkdf(ck, ikm)
	t.Logf("%x", a)
	t.Logf("%x", b)
	t.Logf("%x", c)
}

func TestHandshake(t *testing.T) {
	// generate local static noise key
	kp_init := GenerateKeypair()
	kp_resp := GenerateKeypair()

	libp2p_pub_init,libp2p_priv_init,err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	prologue := []byte("/noise/0.0.0")

	// initiator: new XX noise session
	ns_init := InitSession(true, prologue, kp_init, kp_resp.PubKey())

	// responder: new XX noise session
	ns_resp := InitSession(false, prologue, kp_resp, kp_init.PubKey())

	// stage 0: initiator
	// create payload
	payload_init := new(pb.NoiseHandshakePayload)
	payload_init.Libp2PKey = libp2p_pub_init.Raw()
	payload_init_enc, err := proto.Marshal(payload_init)
	if err != nil {
		t.Fatalf("proto marshal payload fail: %s", err)
	}
	
	// send message
	var msgbuf MessageBuffer
	msg := []byte{}
	msg = append(msg, payload_init_enc[:]...)
	ns_init, msgbuf = SendMessage(ns_init, msg)

	t.Logf("stage 0 msgbuf: %v", msgbuf)

	// stage 0: responder
	var plaintext []byte
	var valid bool
	ns_resp, plaintext, valid = RecvMessage(ns_resp, &msgbuf)
	if !valid {
		t.Fatalf("stage 0 receive not valid")
	}

	t.Logf("stage 0 resp payload: %x", plaintext)

	// stage 1: responder
	// create payload
	payload_resp := new(pb.NoiseHandshakePayload)
	//payload.Libp2PKey()
	payload_resp_enc, err := proto.Marshal(payload_resp)
	if err != nil {
		t.Fatalf("proto marshal payload fail: %s", err)
	}
	msg = append(msg, payload_resp_enc[:]...)
	ns_resp, msgbuf = SendMessage(ns_resp, msg)

	t.Logf("stage 1 msgbuf: %v", msgbuf)

	// stage 1: initiator
	ns_init, plaintext, valid = RecvMessage(ns_init, &msgbuf)
	if !valid {
		t.Fatalf("stage 1 receive not valid")
	} 

	t.Logf("stage 1 resp payload: %x", plaintext)

	// stage 2: initiator
	payload_init = new(pb.NoiseHandshakePayload)
	//payload.Libp2PKey()
	payload_init_enc, err = proto.Marshal(payload_init)
	if err != nil {
		t.Fatalf("proto marshal payload fail: %s", err)
	}
	
	// send message
	msg = append(msg, payload_init_enc[:]...)
	ns_init, msgbuf = SendMessage(ns_init, msg)

	t.Logf("stage 2 msgbuf: %v", msgbuf)

	// stage 2: responder
	ns_resp, plaintext, valid = RecvMessage(ns_resp, &msgbuf)
	if !valid {
		t.Fatalf("stage 2 receive not valid")
	}

	t.Logf("stage 2 resp payload: %x", plaintext)

}