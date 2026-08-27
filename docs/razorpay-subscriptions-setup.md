# Razorpay Subscriptions & Autopay Setup Guide (Test Mode)

This guide walks you through setting up real recurring subscriptions and autopay webhooks on Razorpay Test Dashboard for integration with ReviveOS.

---

## 1. Razorpay Test Mode Configuration

Ensure your [`.env`](file:///c:/Users/jatin/Downloads/Recover/.env) file has your Razorpay Test Key and Secret:
```env
PAYMENT_PROVIDER=razorpay
RAZORPAY_KEY_ID=rzp_test_XXXXXXXXXXXX
RAZORPAY_KEY_SECRET=YYYYYYYYYYYYYYYYYYYY
RAZORPAY_WEBHOOK_SECRET=your_webhook_secret_here
```

---

## 2. Create a Subscription Plan on Razorpay Dashboard

1. Log into your [Razorpay Dashboard](https://dashboard.razorpay.com/) and toggle to **Test Mode**.
2. Navigate to **Subscriptions** &rarr; **Plans** &rarr; **+ Create Plan**.
3. Fill in the plan details:
   - **Plan Name:** e.g. `Pro Monthly Plan`
   - **Billing Frequency:** `Monthly`
   - **Billing Amount:** e.g. `₹1,999.00`
   - **Currency:** `INR`
4. Click **Create Plan** and copy the generated **Plan ID** (e.g., `plan_N0xXXXXXXXXXXX`).

---

## 3. Create a Subscription & Register Customer e-Mandate

1. Navigate to **Subscriptions** &rarr; **Subscriptions** &rarr; **+ Create Subscription**.
2. Select your newly created **Plan**.
3. Enter customer details:
   - **Customer Email:** `subscriber@example.com`
   - **Customer Phone:** `+919876543210`
   - **Total Billing Cycles:** e.g. `12`
   - **Start Date:** `Immediately` or future date.
4. Copy the generated **Subscription Link** (or Subscription ID `sub_XXXXXXXXXXXX`).
5. Open the link in an incognito window and authorize the initial charge / mandate using Razorpay Test Cards (e.g., standard OTP `123456`).

---

## 4. Configure Razorpay Webhooks for Autopay Failures

1. In Razorpay Dashboard, go to **Settings** &rarr; **Webhooks** &rarr; **+ Add New Webhook**.
2. **Webhook URL:**
   - Production / Render: `https://reviveos.onrender.com/api/v1/webhooks/razorpay`
   - Local testing (via ngrok/localtunnel): `https://your-tunnel.ngrok.app/api/v1/webhooks/razorpay`
3. **Secret:** Set the same secret as in `RAZORPAY_WEBHOOK_SECRET`.
4. **Select Events to Subscribe:**
   - `payment.failed` *(Triggers adaptive recovery on any payment decline)*
   - `payment.captured` *(Verifies successful recovery)*
   - `subscription.charged` *(Fires when recurring autopay cycle succeeds or fails)*
   - `subscription.halted` *(Fires when all retries are exhausted)*
   - `invoice.payment_failed` *(Fires on recurring cycle failure)*
5. Click **Save Webhook**.

---

## 5. Simulating Recurring Failures & Observing ReviveOS Recovery

| Test Scenario | Gateway Error Code | Razorpay Card / Trigger | ReviveOS Orchestration |
| :--- | :--- | :--- | :--- |
| **Transient Low Balance** | `INSUFFICIENT_FUNDS` | Simulator decline on charge | **Zero-Touch Automated Token Retry** scheduled after 24h / salary date. |
| **Card / Mandate Expired** | `EXPIRED_CARD` or `MANDATE_FAILED` | Expired test card | **Smart Recovery Email (Resend)** dispatched with update link; subscription set to `PAST_DUE`. |
| **Bank Gateway Down** | `BANK_UNAVAILABLE` | Gateway failure simulator | **Immediate Exponential Jitter Retry** without bothering customer. |

---

## 6. Running the Local Autopay Recovery Test

To test the entire recurring recovery lifecycle without waiting for monthly billing cycles:
```powershell
go run ./apps/worker/cmd/recurring_test/main.go
```
