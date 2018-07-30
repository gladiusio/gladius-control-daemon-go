package peer

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/memberlist"
	uuid "github.com/satori/go.uuid"
)

type mergeDelegate struct {
	peer *Peer
}

// NotifyMerge will be called when a join merge event is invoked. It will
// challenge all nodes in the incoming cluster and verify that they are allowed
// into the network by sending them a challenge that they must sign with their
// Ethereum key.
func (md *mergeDelegate) NotifyMerge(peers []*memberlist.Node) error {
	challengeID := uuid.NewV4().String()
	md.peer.registerOutgoingChallenge(challengeID)

	challengeMap := make(map[string]bool)

	// Go through all nodes in the cluster requesting to join
	for _, peer := range peers {
		// Make a token
		b := make([]byte, 8)
		rand.Read(b)
		questionString := fmt.Sprintf("%x", b)

		// Add it to the map so we can do later lookup to prevent repeat challenges
		challengeMap[questionString] = false

		// Create a challenge from the token
		challenge := &challenge{question: questionString}
		challengeBytes, _ := json.Marshal(challenge)

		// Create an action for the remote peer to process
		action := &update{Action: "challenge_question", Data: challengeBytes}
		actionBytes, _ := json.Marshal(action)

		// Send the message to the remote peer
		md.peer.member.SendReliable(peer, actionBytes)
	}

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(2 * time.Second)
		timeout <- true
	}()

	incomingResponses, err := md.peer.getChallengeResponseChannel(challengeID)
	if err != nil {
		return errors.New("Node responded with unknown challenge ID")
	}
	successfulCount := 0

	// Wait until timeout has completed or a new response has come in
	for {
		select {
		case sm := <-incomingResponses:
			if !sm.IsInPoolAndVerified() {
				return errors.New("Challenge message from peer is not verified or not in pool")
			}
			// Get the challenge response message
			cbytes, err := sm.Message.MarshalJSON()
			if err != nil {
				return errors.New("Can't parse challenge from signed message")
			}
			c := &challenge{}
			err = json.Unmarshal(cbytes, c)
			if err != nil {
				return errors.New("Challenge sent back is corrupted")
			}

			// If the value exists
			if val, ok := challengeMap[c.question]; ok {
				if val {
					return errors.New("Challenged has already been used")
				}
				challengeMap[c.question] = true
				successfulCount++
			}

			// All challenges have successfully been received
			if successfulCount == len(challengeMap) {
				return nil
			}
			// Value must not be in the issued challenges
			return errors.New("Peer returned an unused challenge")

		case <-timeout:
			return errors.New("Timeout reached before all nodes in joining cluster could respond")
		}
	}

}
