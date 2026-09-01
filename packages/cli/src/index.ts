import { scanProject } from "./scanner";
import { generateEnvironmentFile, generateWebhookHandler } from "./generator";
import { runDoctor } from "./doctor";
import { TerminalUI } from "./ui";

export async function runInit(cwd: string = process.cwd()) {
  TerminalUI.printBanner();

  const scan = scanProject(cwd);

  if (!scan.isNodeProject) {
    TerminalUI.error("No Node.js project (package.json) detected in the current directory.");
    TerminalUI.info("Please navigate to your project root and run 'npx reviveos init' again.");
    return;
  }

  TerminalUI.success(`Node.js project detected (${scan.projectName})`);

  if (scan.hasRazorpay) {
    TerminalUI.success("Razorpay integration detected in package.json");
  } else {
    TerminalUI.info("Configuring Razorpay adapter for project...");
  }

  TerminalUI.section("Configuring ReviveOS...");

  const envFile = generateEnvironmentFile(cwd);
  TerminalUI.success("SDK configured");
  TerminalUI.success(`Environment template created: ${envFile}`);

  const webhookPath = generateWebhookHandler(scan);
  if (webhookPath) {
    TerminalUI.success(`Webhook handler generated: ${webhookPath}`);
  }

  TerminalUI.success("Failure events enabled");
  TerminalUI.success("Fraud detection enabled");
  TerminalUI.success("Recovery engine enabled");
  TerminalUI.success("Customer memory enabled");

  TerminalUI.readyBanner();

  TerminalUI.section("Next Steps for Integration:");
  console.log("  1. In Razorpay Dashboard (Settings > Webhooks), set Webhook URL to:");
  console.log("     \x1b[36mhttps://your-domain.com/api/reviveos/webhook\x1b[0m");
  console.log("  2. Subscribe to events: \x1b[33mpayment.failed, payment.captured, refund.created\x1b[0m");
  console.log("  3. Set \x1b[32mRAZORPAY_WEBHOOK_SECRET\x1b[0m in your .env file");
  console.log("  4. Start your server and watch ReviveOS automatically recover revenue!\n");
}

export async function main() {
  const args = process.argv.slice(2);
  const command = args[0] || "init";

  switch (command) {
    case "init":
      await runInit();
      break;
    case "doctor":
      await runDoctor();
      break;
    case "version":
    case "-v":
    case "--version":
      console.log("ReviveOS CLI v1.0.0");
      break;
    default:
      console.log(`Unknown command: ${command}`);
      console.log("Usage: npx reviveos [init | doctor]");
      break;
  }
}
