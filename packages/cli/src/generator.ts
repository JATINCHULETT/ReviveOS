import * as fs from "fs";
import * as path from "path";
import { ProjectScanResult } from "./scanner";

export interface GenerationResult {
  envFileCreated: string;
  webhookFileCreated?: string;
}

export function generateEnvironmentFile(cwd: string): string {
  const envPath = path.join(cwd, ".env.reviveos");
  const envContent = `# =================================================================
# ReviveOS AI Payment Recovery Configuration
# Generated automatically by 'npx reviveos init'
# =================================================================

# Your ReviveOS Merchant API Key (obtained from ReviveOS Dashboard)
REVIVEOS_API_KEY=rvo_test_acme_secret_key_12345

# ReviveOS Core API URL (Defaults to https://api.reviveos.io in production)
REVIVEOS_API_URL=http://localhost:8080

# Razorpay Webhook Secret (configured in Razorpay Dashboard > Settings > Webhooks)
RAZORPAY_WEBHOOK_SECRET=whsec_your_razorpay_webhook_secret

# ReviveOS Execution Mode: 'live', 'test', or 'mock' (for local sandbox testing)
REVIVEOS_MODE=live
`;

  fs.writeFileSync(envPath, envContent, "utf8");

  // Also check if .env exists, if not create standard .env
  const mainEnvPath = path.join(cwd, ".env");
  if (!fs.existsSync(mainEnvPath)) {
    fs.writeFileSync(mainEnvPath, envContent, "utf8");
  }

  return envPath;
}

export function generateWebhookHandler(scan: ProjectScanResult): string | undefined {
  const ext = scan.hasTypeScript ? "ts" : "js";
  let targetPath = "";
  let fileContent = "";

  const isSrcDir = fs.existsSync(path.join(scan.appRoot, "src"));
  const baseDir = isSrcDir ? path.join(scan.appRoot, "src") : scan.appRoot;

  switch (scan.framework) {
    case "nextjs-app": {
      const appDir = fs.existsSync(path.join(scan.appRoot, "src", "app"))
        ? path.join(scan.appRoot, "src", "app")
        : path.join(scan.appRoot, "app");
      const targetDir = path.join(appDir, "api", "reviveos", "webhook");
      fs.mkdirSync(targetDir, { recursive: true });
      targetPath = path.join(targetDir, `route.${ext}`);

      fileContent = `import { NextResponse } from "next/server";
import { ReviveOS } from "@reviveos/razorpay";

// Initialize ReviveOS Recovery Engine
const revive = new ReviveOS({
  apiKey: process.env.REVIVEOS_API_KEY,
  webhookSecret: process.env.RAZORPAY_WEBHOOK_SECRET,
  endpoint: process.env.REVIVEOS_API_URL,
});

export async function POST(req: Request) {
  try {
    const rawBody = await req.text();
    const signature = req.headers.get("x-razorpay-signature") || "";

    // 1. Cryptographically verify signature and normalize payload
    const event = revive.webhooks.verifyAndNormalize(rawBody, signature);

    // 2. Ingest into ReviveOS AI Recovery Engine
    const result = await revive.events.process(event);

    return NextResponse.json(result, { status: 200 });
  } catch (error: any) {
    console.error("[ReviveOS Webhook Error]:", error.message);
    return NextResponse.json(
      { error: error.message || "Failed to process payment event" },
      { status: error.name === "ReviveOSSignatureError" ? 401 : 400 }
    );
  }
}
`;
      break;
    }

    case "nextjs-pages": {
      const pagesDir = fs.existsSync(path.join(scan.appRoot, "src", "pages"))
        ? path.join(scan.appRoot, "src", "pages")
        : path.join(scan.appRoot, "pages");
      const targetDir = path.join(pagesDir, "api", "reviveos");
      fs.mkdirSync(targetDir, { recursive: true });
      targetPath = path.join(targetDir, `webhook.${ext}`);

      fileContent = `import type { NextApiRequest, NextApiResponse } from "next";
import { ReviveOS } from "@reviveos/razorpay";

// Disable Next.js default body parser to get raw text for cryptographic HMAC verification
export const config = {
  api: {
    bodyParser: false,
  },
};

async function getRawBody(req: NextApiRequest): Promise<string> {
  const chunks: Buffer[] = [];
  for await (const chunk of req) {
    chunks.push(typeof chunk === "string" ? Buffer.from(chunk) : chunk);
  }
  return Buffer.concat(chunks).toString("utf8");
}

const revive = new ReviveOS();

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== "POST") {
    return res.status(405).json({ error: "Method not allowed" });
  }

  try {
    const rawBody = await getRawBody(req);
    const signature = (req.headers["x-razorpay-signature"] as string) || "";

    const event = revive.webhooks.verifyAndNormalize(rawBody, signature);
    const result = await revive.events.process(event);

    return res.status(200).json(result);
  } catch (error: any) {
    return res.status(400).json({ error: error.message });
  }
}
`;
      break;
    }

    case "express":
    default: {
      const targetDir = path.join(baseDir, "api");
      fs.mkdirSync(targetDir, { recursive: true });
      targetPath = path.join(targetDir, `reviveos-webhook.${ext}`);

      fileContent = `import { Request, Response } from "express";
import { ReviveOS } from "@reviveos/razorpay";

const revive = new ReviveOS({
  apiKey: process.env.REVIVEOS_API_KEY,
  webhookSecret: process.env.RAZORPAY_WEBHOOK_SECRET,
});

/**
 * Express Route Handler for ReviveOS Razorpay Webhooks.
 * Ensure app.use('/api/reviveos/webhook', express.raw({ type: 'application/json' })) is used before json body parser.
 */
export async function handleReviveOSWebhook(req: Request, res: Response) {
  try {
    const rawBody = req.body;
    const signature = (req.headers["x-razorpay-signature"] as string) || "";

    const event = revive.webhooks.verifyAndNormalize(rawBody, signature);
    const result = await revive.events.process(event);

    return res.status(200).json(result);
  } catch (error: any) {
    return res.status(400).json({ error: error.message });
  }
}
`;
      break;
    }
  }

  fs.writeFileSync(targetPath, fileContent, "utf8");
  return targetPath;
}
