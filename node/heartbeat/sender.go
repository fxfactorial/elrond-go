package heartbeat

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// Sender periodically sends heartbeat messages on a pubsub topic
type Sender struct {
	peerMessenger    PeerMessenger
	singleSigner     crypto.SingleSigner
	privKey          crypto.PrivateKey
	marshalizer      marshal.Marshalizer
	topic            string
	shardCoordinator sharding.Coordinator
	versionNumber    string
	nodeDisplayName  string
}

// NewSender will create a new sender instance
func NewSender(
	peerMessenger PeerMessenger,
	singleSigner crypto.SingleSigner,
	privKey crypto.PrivateKey,
	marshalizer marshal.Marshalizer,
	topic string,
	shardCoordinator sharding.Coordinator,
	versionNumber string,
	nodeDisplayName string,
) (*Sender, error) {
	if check.IfNil(peerMessenger) {
		return nil, ErrNilMessenger
	}
	if check.IfNil(singleSigner) {
		return nil, ErrNilSingleSigner
	}
	if check.IfNil(privKey) {
		return nil, ErrNilPrivateKey
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(shardCoordinator) {
		return nil, ErrNilShardCoordinator
	}

	sender := &Sender{
		peerMessenger:    peerMessenger,
		singleSigner:     singleSigner,
		privKey:          privKey,
		marshalizer:      marshalizer,
		topic:            topic,
		shardCoordinator: shardCoordinator,
		versionNumber:    versionNumber,
		nodeDisplayName:  nodeDisplayName,
	}

	return sender, nil
}

// SendHeartbeat broadcasts a new heartbeat message
func (s *Sender) SendHeartbeat() error {

	hb := &Heartbeat{
		Payload:         []byte(fmt.Sprintf("%v", time.Now())),
		ShardID:         s.shardCoordinator.SelfId(),
		VersionNumber:   s.versionNumber,
		NodeDisplayName: s.nodeDisplayName,
	}

	var err error
	hb.Pubkey, err = s.privKey.GeneratePublic().ToByteArray()
	if err != nil {
		return err
	}

	err = verifyLengths(hb)
	if err != nil {
		log.Warn("verify hb length", "error", err.Error())
		trimLengths(hb)
	}

	hbBytes, err := s.marshalizer.Marshal(hb)
	if err != nil {
		return err
	}

	hb.Signature, err = s.singleSigner.Sign(s.privKey, hbBytes)
	if err != nil {
		return err
	}

	buffToSend, err := s.marshalizer.Marshal(hb)
	if err != nil {
		return err
	}

	s.peerMessenger.Broadcast(s.topic, buffToSend)

	return nil
}
