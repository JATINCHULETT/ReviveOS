import os
import joblib
import numpy as np
from sklearn.ensemble import RandomForestClassifier
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import StandardScaler

MODELS_DIR = os.path.dirname(os.path.abspath(__file__))

def generate_dummy_models():
    os.makedirs(MODELS_DIR, exist_ok=True)
    
    # 1. Fraud Model: Features = [amount, attempt_number, customer_failed_count, customer_success_count, velocity_1h]
    # Synthetic dataset for baseline
    np.random.seed(42)
    X_fraud = np.random.rand(200, 5) * [50000, 5, 10, 10, 5]
    # Probability increases with high amount, high failed count, high velocity
    y_fraud = ((X_fraud[:, 0] > 30000) | (X_fraud[:, 2] > 3) | (X_fraud[:, 4] > 2)).astype(int)
    
    fraud_pipeline = Pipeline([
        ('scaler', StandardScaler()),
        ('classifier', RandomForestClassifier(n_estimators=20, random_state=42))
    ])
    fraud_pipeline.fit(X_fraud, y_fraud)
    
    fraud_path = os.path.join(MODELS_DIR, 'fraud_model.pkl')
    joblib.dump(fraud_pipeline, fraud_path)
    print(f"[ML-Service] Saved dummy fraud model -> {fraud_path}")

    # 2. Return Risk Model: Features = [amount, customer_previous_returns, category_risk_score, days_since_first_order]
    X_return = np.random.rand(200, 4) * [20000, 5, 1.0, 365]
    y_return = ((X_return[:, 0] > 10000) & (X_return[:, 1] > 1)).astype(int)
    
    return_pipeline = Pipeline([
        ('scaler', StandardScaler()),
        ('classifier', RandomForestClassifier(n_estimators=20, random_state=42))
    ])
    return_pipeline.fit(X_return, y_return)
    
    return_path = os.path.join(MODELS_DIR, 'return_model.pkl')
    joblib.dump(return_pipeline, return_path)
    print(f"[ML-Service] Saved dummy return model -> {return_path}")

if __name__ == '__main__':
    generate_dummy_models()
