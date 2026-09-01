#!/usr/bin/env node

const { main } = require("../dist/index");

main().catch((err) => {
  console.error("Fatal ReviveOS CLI error:", err);
  process.exit(1);
});
