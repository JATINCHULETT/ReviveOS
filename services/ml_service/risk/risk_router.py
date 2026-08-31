import logging

try:
    from .schemas import RiskAnalysisRequest, RiskAnalysisResponse, RiskScore
    from ..predictors.fraud_predictor import FraudPredictor
    from ..predictors.return_predictor import ReturnPredictor
except ImportError:
    from risk.schemas import RiskAnalysisRequest, RiskAnalysisResponse, RiskScore
    from predictors.fraud_predictor import FraudPredictor
    from predictors.return_predictor import ReturnPredictor

logger = logging.getLogger(__name__)

class RiskRouter:
    def __init__(self):
        self.fraud_predictor = FraudPredictor()
        self.return_predictor = ReturnPredictor()

    def analyze(self, req: RiskAnalysisRequest) -> RiskAnalysisResponse:
        event = req.event_type.upper()
        amount = req.amount
        features = req.model_dump()

        # 1. Event-Aware Routing
        if event in ["PAYMENT_FAILED", "PAYMENT.FAILED", "FAILED"]:
            # Route strictly to Fraud Model
            fraud_res = self.fraud_predictor.predict(features)
            fraud_score = RiskScore(**fraud_res)
            return_score = None

            # Determine action based on fraud
            if fraud_score.probability >= 0.70:
                overall_risk = "HIGH"
                expected_loss = round(amount * fraud_score.probability, 2)
                recommended_action = "VERIFY_FRAUD_ESCALATE"
                reason = f"High fraud risk ({fraud_score.probability*100:.1f}%) detected on payment failure. Automated retry halted."
            elif fraud_score.probability >= 0.35:
                overall_risk = "MEDIUM"
                expected_loss = round(amount * fraud_score.probability * 0.5, 2)
                recommended_action = "PROCEED_WITH_GUARDRAILS"
                reason = f"Moderate fraud risk ({fraud_score.probability*100:.1f}%). Safe to recover with standard policies."
            else:
                overall_risk = "LOW"
                expected_loss = 0.0
                recommended_action = "ALLOW_AUTONOMOUS_RECOVERY"
                reason = f"Low fraud risk ({fraud_score.probability*100:.1f}%). Verified safe for autonomous recovery."

        elif event in ["PAYMENT_SUCCESS", "PAYMENT.SUCCESS", "CAPTURED", "RECOVERED", "SUCCESS"]:
            # Route to Fraud AND Return Risk Models
            fraud_res = self.fraud_predictor.predict(features)
            fraud_score = RiskScore(**fraud_res)

            return_res = self.return_predictor.predict(features)
            return_score = RiskScore(**return_res)

            max_prob = max(fraud_score.probability, return_score.probability)
            if max_prob >= 0.60:
                overall_risk = "HIGH"
                expected_loss = round(amount * return_score.probability, 2)
                recommended_action = "MONITOR_POST_PAYMENT_RISK"
                reason = f"Payment settled, but elevated return risk ({return_score.probability*100:.1f}%) flagged for fulfillment monitoring."
            elif max_prob >= 0.30:
                overall_risk = "MEDIUM"
                expected_loss = round(amount * return_score.probability * 0.4, 2)
                recommended_action = "STANDARD_FULFILLMENT"
                reason = f"Payment settled with normal baseline return risk ({return_score.probability*100:.1f}%)."
            else:
                overall_risk = "LOW"
                expected_loss = 0.0
                recommended_action = "SAFE_SETTLEMENT"
                reason = "Payment verified safe across both fraud and return vectors."

        else:
            # Generic fallback
            fraud_res = self.fraud_predictor.predict(features)
            fraud_score = RiskScore(**fraud_res)
            return_score = None
            overall_risk = fraud_score.risk_level
            expected_loss = round(amount * fraud_score.probability, 2)
            recommended_action = "ALLOW"
            reason = "Evaluated generic transaction risk."

        return RiskAnalysisResponse(
            event_type=req.event_type,
            payment_id=req.payment_id,
            fraud=fraud_score,
            return_risk=return_score,
            overall_risk=overall_risk,
            expected_loss=expected_loss,
            recommended_action=recommended_action,
            reason=reason
        )
