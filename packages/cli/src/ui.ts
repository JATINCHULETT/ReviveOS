export class TerminalUI {
  public static printBanner() {
    console.log("\n\x1b[36m╭──────────────────────────────────────────────────╮");
    console.log("│                  \x1b[1m\x1b[37mREVIVEOS AI\x1b[0m\x1b[36m                     │");
    console.log("│       \x1b[33mAI Payment Recovery & Revenue Layer\x1b[0m\x1b[36m        │");
    console.log("╰──────────────────────────────────────────────────╯\x1b[0m\n");
  }

  public static success(message: string) {
    console.log(`\x1b[32m✓\x1b[0m ${message}`);
  }

  public static info(message: string) {
    console.log(`\x1b[34mℹ\x1b[0m ${message}`);
  }

  public static warn(message: string) {
    console.log(`\x1b[33m⚠\x1b[0m ${message}`);
  }

  public static error(message: string) {
    console.log(`\x1b[31m✗\x1b[0m ${message}`);
  }

  public static section(title: string) {
    console.log(`\n\x1b[1m${title}\x1b[0m`);
  }

  public static readyBanner() {
    console.log("\n\x1b[32m\x1b[1mREVIVEOS IS READY ✓\x1b[0m\n");
  }
}
