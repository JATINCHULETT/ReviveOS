package receivables

import (
	"fmt"
	"time"
)

// Bucket represents the aging bucket for an invoice
type Bucket string

const (
	BucketCurrent  Bucket = "CURRENT"
	Bucket1To30    Bucket = "1_30"
	Bucket31To60   Bucket = "31_60"
	Bucket61To90   Bucket = "61_90"
	Bucket90Plus   Bucket = "90_PLUS"
)

// Invoice represents a B2B trade invoice
type Invoice struct {
	ID             string     `json:"id"`
	MerchantID     string     `json:"merchant_id"`
	InvoiceNumber  string     `json:"invoice_number"`
	BuyerCompany   string     `json:"buyer_company"`
	BuyerName      string     `json:"buyer_name"`
	BuyerEmail     string     `json:"buyer_email"`
	BuyerPhone     string     `json:"buyer_phone,omitempty"`
	Amount         float64    `json:"amount"`
	Currency       string     `json:"currency"`
	IssueDate      time.Time  `json:"issue_date"`
	DueDate        time.Time  `json:"due_date"`
	Status         string     `json:"status"` // UNPAID, PARTIALLY_PAID, PAID, OVERDUE, WRITTEN_OFF
	CreditTerms    string     `json:"credit_terms"`
	CurrentBucket  Bucket     `json:"current_bucket"`
	DaysPastDue    int        `json:"days_past_due"`
	DunningStage   int        `json:"dunning_stage"`
	LastContactAt  *time.Time `json:"last_contacted_at,omitempty"`
	NextActionDue  *time.Time `json:"next_action_due,omitempty"`
	PaymentLinkURL string     `json:"payment_link_url,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// DunningRecommendation details recommended automated action
type DunningRecommendation struct {
	Stage          int        `json:"stage"`
	Channel        string     `json:"channel"` // EMAIL, SMS, VOICE, WHATSAPP
	Tone           string     `json:"tone"`    // COURTEOUS, FIRM, URGENT, LEGAL_NOTICE
	TemplateSubject string    `json:"template_subject"`
	TemplateBody   string     `json:"template_body"`
	SendAfter      time.Time  `json:"send_after"`
	GeneratePaymentLink bool  `json:"generate_payment_link"`
}

// CalculateAging computes days past due and aging bucket
func CalculateAging(dueDate time.Time, asOf time.Time) (int, Bucket) {
	if asOf.Before(dueDate) {
		return 0, BucketCurrent
	}
	diffDays := int(asOf.Sub(dueDate).Hours() / 24)
	if diffDays <= 0 {
		return 0, BucketCurrent
	} else if diffDays <= 30 {
		return diffDays, Bucket1To30
	} else if diffDays <= 60 {
		return diffDays, Bucket31To60
	} else if diffDays <= 90 {
		return diffDays, Bucket61To90
	}
	return diffDays, Bucket90Plus
}

// NextDunningAction determines the dunning escalation step based on overdue status
func NextDunningAction(inv Invoice, asOf time.Time) DunningRecommendation {
	days, bucket := CalculateAging(inv.DueDate, asOf)

	switch bucket {
	case BucketCurrent:
		return DunningRecommendation{
			Stage: 0,
			Channel: "EMAIL",
			Tone: "COURTEOUS",
			TemplateSubject: fmt.Sprintf("Upcoming Due Date: Invoice #%s from %s", inv.InvoiceNumber, inv.BuyerCompany),
			TemplateBody: fmt.Sprintf("Hello %s,\n\nThis is a friendly reminder that Invoice #%s for %.2f %s is due on %s.",
				inv.BuyerName, inv.InvoiceNumber, inv.Amount, inv.Currency, inv.DueDate.Format("02 Jan 2006")),
			SendAfter: inv.DueDate.Add(-48 * time.Hour),
			GeneratePaymentLink: true,
		}
	case Bucket1To30:
		return DunningRecommendation{
			Stage: 1,
			Channel: "EMAIL",
			Tone: "FIRM",
			TemplateSubject: fmt.Sprintf("Overdue Notice: Invoice #%s (%d Days Past Due)", inv.InvoiceNumber, days),
			TemplateBody: fmt.Sprintf("Dear %s,\n\nInvoice #%s for %.2f %s was due on %s and is now %d days past due. Please settle via the payment link below.",
				inv.BuyerName, inv.InvoiceNumber, inv.Amount, inv.Currency, inv.DueDate.Format("02 Jan 2006"), days),
			SendAfter: asOf,
			GeneratePaymentLink: true,
		}
	case Bucket31To60:
		return DunningRecommendation{
			Stage: 2,
			Channel: "WHATSAPP",
			Tone: "URGENT",
			TemplateSubject: fmt.Sprintf("Urgent Payment Required: Invoice #%s", inv.InvoiceNumber),
			TemplateBody: fmt.Sprintf("Attention %s: Invoice #%s (%.2f %s) is %d days overdue. ReviveOS AI Receivables has flagged your account for review.",
				inv.BuyerName, inv.InvoiceNumber, inv.Amount, inv.Currency, days),
			SendAfter: asOf,
			GeneratePaymentLink: true,
		}
	default: // 61+
		return DunningRecommendation{
			Stage: 3,
			Channel: "VOICE",
			Tone: "LEGAL_NOTICE",
			TemplateSubject: fmt.Sprintf("Final Notice Before Escalation: Invoice #%s", inv.InvoiceNumber),
			TemplateBody: fmt.Sprintf("Final escalation notice for %s regarding unpaid balance of %.2f %s.",
				inv.BuyerCompany, inv.Amount, inv.Currency),
			SendAfter: asOf,
			GeneratePaymentLink: true,
		}
	}
}
