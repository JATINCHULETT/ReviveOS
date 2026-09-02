package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/reviveos/api/handlers"
	"github.com/reviveos/checkout"
	"github.com/reviveos/ptp"
	"github.com/reviveos/receivables"
	"github.com/reviveos/voice"
)

func main() {
	fmt.Println("===============================================================")
	fmt.Println("     REVIVEOS EXPANDED SUITE VERIFICATION & TEST RUNNER")
	fmt.Println("===============================================================")

	// 1. Receivables Unit & Integration
	fmt.Println("\n[1/4] Testing B2B Receivables Chaser...")
	now := time.Now()
	inv := receivables.Invoice{
		InvoiceNumber: "INV-VERIFY-001",
		BuyerCompany:  "Tata Consultancy Test",
		BuyerName:     "Amit Verma",
		BuyerEmail:    "amit@tataconsultancy.test",
		Amount:        75000,
		Currency:      "INR",
		DueDate:       now.Add(-25 * 24 * time.Hour), // 25 days overdue
	}
	days, bucket := receivables.CalculateAging(inv.DueDate, now)
	rec := receivables.NextDunningAction(inv, now)
	if bucket != receivables.Bucket1To30 || days != 25 {
		panic(fmt.Sprintf("Failed aging calculation: got %s, days %d", bucket, days))
	}
	if rec.Stage != 1 || rec.Tone != "FIRM" {
		panic("Failed dunning recommendation verification")
	}
	fmt.Printf("   Aging correctly evaluated: 25 days past due -> Bucket: %s\n", bucket)
	fmt.Printf("   Dunning Action: Stage %d (%s) via Resend Email\n", rec.Stage, rec.Tone)

	// 2. Hinglish Voice Engine Unit & Integration
	fmt.Println("\n[2/4] Testing Hinglish AI Voice Recovery Engine...")
	callPayload := voice.CallPayload{
		CustomerName: "Rajesh",
		Phone:        "+919876543210",
		Amount:       12999,
		Currency:     "INR",
		DueDate:      now,
	}
	script := voice.GenerateHinglishScript(callPayload)
	intent, ptpDate := voice.ClassifyHinglishIntent("Namaste ji, main kal subah tak payment kar dunga")
	if intent != voice.IntentPromiseToPay || ptpDate == nil {
		panic("Failed voice intent classification for PROMISE_TO_PAY")
	}
	fmt.Printf("   Hinglish Script Generated: %s\n", script[:70]+"...")
	fmt.Printf("   Intent Extracted: %s (Scheduled for: %s)\n", intent, ptpDate.Format("02 Jan 15:04"))

	// 3. Promise-to-Pay Tracker State Machine
	fmt.Println("\n[3/4] Testing Promise-to-Pay (PTP) Tracker...")
	promise := ptp.PromiseRecord{
		PromisedAmount: 12999,
		PromisedDate:   now.Add(24 * time.Hour),
		Status:         ptp.StatusPending,
	}
	err := ptp.ExtendPromise(&promise, now.Add(72*time.Hour), 2)
	if err != nil || promise.Status != ptp.StatusExtended {
		panic("Failed promise extension test")
	}
	status, _ := ptp.EvaluateStatus(promise, now, true)
	if status != ptp.StatusHonored {
		panic("Failed promise reconciliation test")
	}
	fmt.Println("   Commitment registered -> Extended (+3 days) -> Verified Paid (HONORED)")

	// 4. Checkout Drop-off Recovery
	fmt.Println("\n[4/4] Testing Checkout Drop-off Recovery & Cart Telemetry...")
	tok := checkout.GenerateSessionToken()
	isDropped := checkout.IsDroppedOff(now.Add(-30*time.Minute), 15)
	if !isDropped {
		panic("Failed drop-off threshold evaluation")
	}
	recoveryLink := checkout.GenerateRecoveryLink("https://reviveos.onrender.com", tok)
	fmt.Printf("   Session Token Generated: %s\n", tok)
	fmt.Printf("   Abandoned Cart Detected (Idle > 15m) -> 1-Click Restore Link: %s\n", recoveryLink)

	// 5. HTTP Handler Routing Sanity Check
	fmt.Println("\n[5/5] Testing HTTP REST Endpoints...")
	mux := handlers.RegisterRoutes(nil) // nil pool uses fallback / mock data gracefully
	ts := httptest.NewServer(mux)
	defer ts.Close()

	endpoints := []string{
		"/health",
		"/v1/receivables",
		"/v1/voice/scripts/preview",
		"/v1/ptp",
		"/v1/checkout",
	}

	for _, ep := range endpoints {
		resp, err := http.Get(ts.URL + ep)
		if err != nil {
			panic(fmt.Sprintf("Failed GET %s: %v", ep, err))
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			panic(fmt.Sprintf("Endpoint %s returned status %d: %s", ep, resp.StatusCode, string(body)))
		}
		fmt.Printf("   HTTP 200 OK -> %s\n", ep)
	}

	fmt.Println("\n===============================================================")
	fmt.Println("  ALL NEW MODULES & NON-BREAKING VERIFICATIONS PASSED 100%!")
	fmt.Println("===============================================================")
}
