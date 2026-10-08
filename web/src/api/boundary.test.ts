import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative, sep } from "node:path";

const SRC = join(__dirname, "..");
const forbidden = [/\bfetch\s*\(/, /from\s+["']axios["']/, /new\s+XMLHttpRequest\b/];

function files(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const path = join(dir, name);
    return statSync(path).isDirectory() ? files(path) : [path];
  });
}

describe("backend access", () => {
  it("only generated client calls the backend", () => {
    const offenders = files(SRC)
      .filter((f) => /\.(ts|tsx)$/.test(f) && !/\.test\.tsx?$/.test(f))
      .map((f) => relative(SRC, f))
      .filter((f) => !f.startsWith(`api${sep}`) && !f.startsWith(`test${sep}`))
      .filter((f) => forbidden.some((re) => re.test(readFileSync(join(SRC, f), "utf8"))));
    expect(offenders).toEqual([]);
  });
});
