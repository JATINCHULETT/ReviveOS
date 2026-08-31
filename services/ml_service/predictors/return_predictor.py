import os
import joblib
import numpy as np
import logging

logger = logging.getLogger(__name__)

class ReturnPredictor:
    def __init__(self, model_path: str = None):
        if model_path is None:
            model_path = os.path.join(os.path.dirname(__file__), "..", "models", "return_model.pkl")
        self.model_path = os.path.abspath(model_path)
        self.model = None
        self.model_version = "return-rf-v1.0"
        self._load_model()

    def _load_model(self):
        try:
            if os.path.exists(self.model_path):
                self.model = joblib.load(self.model_path)
                logger.info(f"Loaded Return Risk Model from {self.model_path}")
            else:
                logger.warning(f"Return model file not found at {self.model_path}, using heuristic fallback.")
        except Exception as e:
            logger.error(f"Failed to load return model: {e}")
            self.model = None

    def predict(self, features: dict) -> dict:
        """
        Features expected:
        - amount: float
        - customer_previous_returns: int
        - category_risk_score: float (0.0 to 1.0)
        - days_since_first_order: int
        """
        amount = float(features.get("amount", 0.0))
        prev_returns = int(features.get("customer_previous_returns", 0))
        cat_risk = float(features.get("category_risk_score", 0.1))
        days_active = int(features.get("days_since_first_order", 30))

        if self.model is not None:
            try:
                X = np.array([[amount, prev_returns, cat_risk, days_active]])
                if hasattr(self.model, "predict_proba"):
                    probs = self.model.predict_proba(X)
                    return_prob = float(probs[0][1]) if probs.shape[1] > 1 else float(probs[0][0])
                else:
                    pred = self.model.predict(X)
                    return_prob = float(pred[0])
            except Exception as e:
                logger.error(f"Error predicting return risk: {e}")
                return_prob = self._heuristic(amount, prev_returns, cat_risk)
        else:
            return_prob = self._heuristic(amount, prev_returns, cat_risk)

        if return_prob >= 0.60:
            risk_level = "HIGH"
        elif return_prob >= 0.30:
            risk_level = "MEDIUM"
        else:
            risk_level = "LOW"

        return {
            "probability": round(return_prob, 4),
            "risk_level": risk_level,
            "model_version": self.model_version
        }

    def _heuristic(self, amount: float, prev_returns: int, cat_risk: float) -> float:
        score = 0.08
        if prev_returns > 2:
            score += 0.40
        elif prev_returns > 0:
            score += 0.15
        if amount > 15000:
            score += 0.25
        score += cat_risk * 0.30
        return min(0.95, score)
