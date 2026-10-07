// Fails when src/api/schema.d.ts differs from what `npm run gen` produces from app/openapi.json.
// Usage: node scripts/gen-check.mjs [spec] [committed-schema]
import { execFileSync } from "node:child_process";
import { mkdtempSync, readFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const web = join(dirname(fileURLToPath(import.meta.url)), "..");
const spec = process.argv[2] ?? join(web, "..", "app", "openapi.json");
const committed = process.argv[3] ?? join(web, "src", "api", "schema.d.ts");
const tmp = mkdtempSync(join(tmpdir(), "gen-check-"));
const out = join(tmp, "schema.d.ts");
// Run the packages' JS entry points with this node, so no shell is needed on Windows.
const cli = {
  openapiTypescript: join(web, "node_modules", "openapi-typescript", "bin", "cli.js"),
  biome: join(web, "node_modules", "@biomejs", "biome", "bin", "biome"),
};
const opts = { cwd: web, stdio: "pipe" };

try {
  execFileSync(process.execPath, [cli.openapiTypescript, spec, "-o", out], opts);
  execFileSync(process.execPath, [cli.biome, "format", "--write", out], opts);
  if (readFileSync(out, "utf8") !== readFileSync(committed, "utf8")) {
    console.error(`${committed} is out of date with ${spec}: run "npm run gen"`);
    process.exitCode = 1;
  }
} finally {
  rmSync(tmp, { recursive: true, force: true });
}
