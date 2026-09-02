package checkout_test

import (
	"strings"
	"testing"
	"time"

	"github.com/reviveos/checkout"
)

func TestGenerateSessionToken(t *testing.T) {
	token := checkout.GenerateSessionToken()
	if len(token) != 32 {
		t.Errorf("Expected 32 character hex token, got length %d", len(token))
	}
}

func TestIsDroppedOff(t *testing.T) {
	activeTime := time.Now().Add(-5 * time.Minute)
	if checkout.IsDroppedOff(activeTime, 15) {
		t.Errorf("Expected 5 min old session to NOT be dropped off")
	}

	idleTime := time.Now().Add(-25 * time.Minute)
	if !checkout.IsDroppedOff(idleTime, 15) {
		t.Errorf("Expected 25 min old session to be dropped off")
	}
}

func TestGenerateRecoveryLink(t *testing.T) {
	link := checkout.GenerateRecoveryLink("https://reviveos.onrender.com", "tok_xyz_123")
	if !strings.Contains(link, "token=tok_xyz_123") {
		t.Errorf("Expected link to contain token, got %s", link)
	}
}
