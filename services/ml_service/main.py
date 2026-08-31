import logging
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
try:
    from .risk.schemas import RiskAnalysisRequest, RiskAnalysisResponse
    from .risk.risk_router import RiskRouter
except ImportError:
    from risk.schemas import RiskAnalysisRequest, RiskAnalysisResponse
    from risk.risk_router import RiskRouter

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(name)s: %(message)s")
logger = logging.getLogger("ml-service")

app = FastAPI(
    title="ReviveOS ML Risk Service",
    description="Machine Learning Risk Assessment for Fraud Detection & Return Risk Intelligence",
    version="1.0.0"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

risk_router = RiskRouter()

@app.get("/health")
def health_check():
    return {
        "status": "HEALTHY",
        "service": "ReviveOS ML Risk Service",
        "models": {
            "fraud_model": risk_router.fraud_predictor.model is not None,
            "return_model": risk_router.return_predictor.model is not None
        }
    }

@app.post("/risk/analyze", response_model=RiskAnalysisResponse)
def analyze_risk(request: RiskAnalysisRequest):
    try:
        res = risk_router.analyze(request)
        logger.info(f"Risk analyzed for payment {request.payment_id} ({request.event_type}): Overall={res.overall_risk}, Fraud={res.fraud.probability}")
        return res
    except Exception as e:
        logger.error(f"Error in analyze_risk: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    import uvicorn
    uvicorn.run("services.ml_service.main:app", host="0.0.0.0", port=8001, reload=True)
