import * as http from "http";
import * as https from "https";
import { URL } from "url";
import { TerminalUI } from "./ui";

export async function checkApiConnectivity(endpoint: string = "http://localhost:8080"): Promise<boolean> {
  return new Promise((resolve) => {
    try {
      const url = new URL(`${endpoint.replace(/\/$/, "")}/health`);
      const client = url.protocol === "https:" ? https : http;

      const req = client.get(url, { timeout: 3000 }, (res) => {
        if (res.statusCode && res.statusCode >= 200 && res.statusCode < 400) {
          resolve(true);
        } else {
          resolve(false);
        }
      });

      req.on("timeout", () => {
        req.destroy();
        resolve(false);
      });

      req.on("error", () => {
        resolve(false);
      });
    } catch {
      resolve(false);
    }
  });
}

export async function runDoctor(cwd: string = process.cwd()) {
  TerminalUI.printBanner();
  TerminalUI.section("Diagnostics & Connectivity Check");

  const endpoint = process.env.REVIVEOS_API_URL || "http://localhost:8080";
  const isReachable = await checkApiConnectivity(endpoint);

  if (isReachable) {
    TerminalUI.success(`ReviveOS API reachable at ${endpoint}`);
  } else {
    TerminalUI.warn(`ReviveOS API not reachable at ${endpoint} (Using sandbox / mock mode fallback)`);
  }

  if (process.env.REVIVEOS_API_KEY) {
    TerminalUI.success("REVIVEOS_API_KEY is configured in environment");
  } else {
    TerminalUI.warn("REVIVEOS_API_KEY is not set (Check .env or .env.reviveos)");
  }

  if (process.env.RAZORPAY_WEBHOOK_SECRET) {
    TerminalUI.success("RAZORPAY_WEBHOOK_SECRET is configured");
  } else {
    TerminalUI.info("RAZORPAY_WEBHOOK_SECRET can be set for strict HMAC-SHA256 verification");
  }
}
