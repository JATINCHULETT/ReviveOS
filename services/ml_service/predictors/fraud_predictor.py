import os
import joblib
import numpy as np
import logging

logger = logging.getLogger(__name__)

class FraudPredictor:
    def __init__(self, model_path: str = None):
        if model_path is None:
            model_path = os.path.join(os.path.dirname(__file__), "..", "models", "fraud_model.pkl")
        self.model_path = os.path.abspath(model_path)
        self.model = None
        self.model_version = "fraud-rf-v1.0"
        self._load_model()

    def _load_model(self):
        try:
            if os.path.exists(self.model_path):
                self.model = joblib.load(self.model_path)
                logger.info(f"Loaded Fraud Model from {self.model_path}")
            else:
                logger.warning(f"Fraud model file not found at {self.model_path}, using heuristic fallback.")
        except Exception as e:
            logger.error(f"Failed to load fraud model: {e}")
            self.model = None

    def predict(self, features: dict) -> dict:
        """
        Features expected:
        - amount: float
        - attempt_number: int
        - customer_failed_count: int
        - customer_success_count: int
        - velocity_1h: int
        """
        amount = float(features.get("amount", 0.0))
        attempt = int(features.get("attempt_number", 1))
        failed_count = int(features.get("customer_failed_count", 0))
        success_count = int(features.get("customer_success_count", 0))
        velocity = int(features.get("velocity_1h", 1))

        if self.model is not None:
            try:
                X = np.array([[amount, attempt, failed_count, success_count, velocity]])
                if hasattr(self.model, "predict_proba"):
                    probs = self.model.predict_proba(X)
                    fraud_prob = float(probs[0][1]) if probs.shape[1] > 1 else float(probs[0][0])
                else:
                    pred = self.model.predict(X)
                    fraud_prob = float(pred[0])
            except Exception as e:
                logger.error(f"Error predicting fraud: {e}")
                fraud_prob = self._heuristic(amount, attempt, failed_count, velocity)
        else:
            fraud_prob = self._heuristic(amount, attempt, failed_count, velocity)

        # Risk level determination
        if fraud_prob >= 0.70:
            risk_level = "HIGH"
        elif fraud_prob >= 0.35:
            risk_level = "MEDIUM"
        else:
            risk_level = "LOW"

        return {
            "probability": round(fraud_prob, 4),
            "risk_level": risk_level,
            "model_version": self.model_version
        }

    def _heuristic(self, amount: float, attempt: int, failed_count: int, velocity: int) -> float:
        score = 0.05
        if amount > 50000:
            score += 0.40
        elif amount > 20000:
            score += 0.20
        if failed_count > 3:
            score += 0.30
        if velocity > 2:
            score += 0.25
        if attempt > 3:
            score += 0.15
        return min(0.95, score)
