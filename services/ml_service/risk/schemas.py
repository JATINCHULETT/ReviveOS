from typing import Optional, Dict, Any
from pydantic import BaseModel, Field

class RiskAnalysisRequest(BaseModel):
    event_type: str = Field(..., description="PAYMENT_FAILED, PAYMENT_SUCCESS, ORDER_CREATED, etc.")
    payment_id: Optional[str] = None
    merchant_id: Optional[str] = None
    customer_id: Optional[str] = None
    customer_email: Optional[str] = None
    amount: float = 0.0
    currency: str = "INR"
    failure_code: Optional[str] = None
    attempt_number: int = 1
    customer_failed_count: int = 0
    customer_success_count: int = 0
    customer_previous_returns: int = 0
    velocity_1h: int = 1
    metadata: Optional[Dict[str, Any]] = None

class RiskScore(BaseModel):
    probability: float
    risk_level: str
    model_version: str

class RiskAnalysisResponse(BaseModel):
    event_type: str
    payment_id: Optional[str] = None
    fraud: RiskScore
    return_risk: Optional[RiskScore] = None
    overall_risk: str
    expected_loss: float
    recommended_action: str
    reason: str
