import * as fs from "fs";
import * as path from "path";

export interface ProjectScanResult {
  isNodeProject: boolean;
  packageJsonPath?: string;
  projectName?: string;
  hasRazorpay: boolean;
  hasTypeScript: boolean;
  framework: "nextjs-app" | "nextjs-pages" | "express" | "fastify" | "node";
  appRoot: string;
}

export function scanProject(cwd: string = process.cwd()): ProjectScanResult {
  const packageJsonPath = path.join(cwd, "package.json");
  const isNodeProject = fs.existsSync(packageJsonPath);

  if (!isNodeProject) {
    return {
      isNodeProject: false,
      hasRazorpay: false,
      hasTypeScript: false,
      framework: "node",
      appRoot: cwd,
    };
  }

  let pkg: any = {};
  try {
    const raw = fs.readFileSync(packageJsonPath, "utf8");
    pkg = JSON.parse(raw);
  } catch {
    // ignore parse error
  }

  const allDeps = {
    ...(pkg.dependencies || {}),
    ...(pkg.devDependencies || {}),
  };

  const hasRazorpay = Boolean(allDeps["razorpay"] || allDeps["@types/razorpay"]);
  const hasTypeScript = Boolean(
    allDeps["typescript"] || fs.existsSync(path.join(cwd, "tsconfig.json"))
  );

  let framework: "nextjs-app" | "nextjs-pages" | "express" | "fastify" | "node" = "node";

  // Check for Next.js
  if (allDeps["next"] || fs.existsSync(path.join(cwd, "next.config.js")) || fs.existsSync(path.join(cwd, "next.config.mjs"))) {
    const hasAppDir =
      fs.existsSync(path.join(cwd, "app")) || fs.existsSync(path.join(cwd, "src", "app"));
    if (hasAppDir) {
      framework = "nextjs-app";
    } else {
      framework = "nextjs-pages";
    }
  } else if (allDeps["express"]) {
    framework = "express";
  } else if (allDeps["fastify"]) {
    framework = "fastify";
  }

  return {
    isNodeProject: true,
    packageJsonPath,
    projectName: pkg.name || path.basename(cwd),
    hasRazorpay,
    hasTypeScript,
    framework,
    appRoot: cwd,
  };
}
