const test = require("node:test");
const assert = require("node:assert");
const fs = require("node:fs");
const path = require("node:path");
const os = require("node:os");

const { scanProject } = require("../dist/scanner");
const { generateEnvironmentFile, generateWebhookHandler } = require("../dist/generator");

test("CLI: Scanner correctly detects Node.js and Razorpay project", () => {
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "reviveos-test-"));
  
  // Setup sample package.json
  const samplePkg = {
    name: "my-ecommerce-store",
    dependencies: {
      next: "^14.0.0",
      razorpay: "^2.9.2",
    },
  };
  fs.writeFileSync(path.join(tmpDir, "package.json"), JSON.stringify(samplePkg, null, 2));
  fs.mkdirSync(path.join(tmpDir, "app"), { recursive: true });

  const result = scanProject(tmpDir);

  assert.strictEqual(result.isNodeProject, true);
  assert.strictEqual(result.projectName, "my-ecommerce-store");
  assert.strictEqual(result.hasRazorpay, true);
  assert.strictEqual(result.framework, "nextjs-app");

  // Clean up
  fs.rmSync(tmpDir, { recursive: true, force: true });
});

test("CLI: Generator creates .env.reviveos and Next.js App Router Webhook", () => {
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "reviveos-gen-"));
  
  const samplePkg = {
    name: "test-app",
    dependencies: { next: "14.0.0" },
  };
  fs.writeFileSync(path.join(tmpDir, "package.json"), JSON.stringify(samplePkg));
  fs.mkdirSync(path.join(tmpDir, "app"), { recursive: true });

  const scan = scanProject(tmpDir);

  // 1. Generate Env
  const envPath = generateEnvironmentFile(tmpDir);
  assert.ok(fs.existsSync(envPath));
  const envContent = fs.readFileSync(envPath, "utf8");
  assert.ok(envContent.includes("REVIVEOS_API_KEY"));

  // 2. Generate Webhook
  const webhookPath = generateWebhookHandler(scan);
  assert.ok(webhookPath);
  assert.ok(fs.existsSync(webhookPath));
  const webhookContent = fs.readFileSync(webhookPath, "utf8");
  assert.ok(webhookContent.includes("@reviveos/razorpay"));
  assert.ok(webhookContent.includes("verifyAndNormalize"));

  // Clean up
  fs.rmSync(tmpDir, { recursive: true, force: true });
});
