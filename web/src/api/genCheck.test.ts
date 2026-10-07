import { spawnSync } from "node:child_process";
import { appendFileSync, copyFileSync, mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const web = join(__dirname, "..", "..");
const spec = join(web, "..", "app", "openapi.json");
const run = (...args: string[]) =>
  spawnSync(process.execPath, [join(web, "scripts", "gen-check.mjs"), ...args], {
    cwd: web,
    encoding: "utf8",
  });

describe("gen:check", () => {
  // C29
  it("drift fails", () => {
    const edited = join(mkdtempSync(join(tmpdir(), "drift-")), "schema.d.ts");
    copyFileSync(join(web, "src", "api", "schema.d.ts"), edited);
    appendFileSync(edited, "\nexport type HandEdited = 1;\n");
    expect(run(spec, edited).status).not.toBe(0);
  }, 30_000);

  // C29
  it("committed schema passes", () => {
    const res = run();
    expect(res.stderr).toBe("");
    expect(res.status).toBe(0);
  }, 30_000);
});
