import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { json, renderAt, stubFetch } from "@/test/render";

const ready = json(200, { status: "ready" });
const down = json(503, {
  type: "about:blank",
  title: "Service Unavailable",
  status: 503,
  detail: "database unavailable",
  request_id: "r-1",
});
const networkError = () => Promise.reject(new TypeError("Failed to fetch"));
const readyz = (urls: string[]) => urls.filter((u) => u.endsWith("/readyz"));

describe("ApiStatus on /", () => {
  it("shows online when readyz answers 200", async () => {
    const calls = stubFetch(ready);
    renderAt("/");
    expect(await screen.findByText("API: online")).toBeInTheDocument();
    expect(readyz(calls)).toHaveLength(1);
  });

  it("shows offline with a retry button when readyz answers 503", async () => {
    stubFetch(down);
    renderAt("/");
    expect(await screen.findByText("API: offline")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Tentar novamente" })).toBeInTheDocument();
  });

  it("shows offline with a retry button when readyz fails at the network", async () => {
    stubFetch(networkError);
    renderAt("/");
    expect(await screen.findByText("API: offline")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Tentar novamente" })).toBeInTheDocument();
  });

  it("retry requests readyz again", async () => {
    const calls = stubFetch(down, ready);
    renderAt("/");
    await userEvent.click(await screen.findByRole("button", { name: "Tentar novamente" }));
    expect(await screen.findByText("API: online")).toBeInTheDocument();
    expect(readyz(calls)).toHaveLength(2);
  });

  it("shows a status element while readyz is pending", async () => {
    let resolve: (r: Response) => void = () => {};
    stubFetch(
      () =>
        new Promise<Response>((r) => {
          resolve = r;
        }),
    );
    renderAt("/");
    expect(await screen.findByRole("status")).toBeInTheDocument();
    resolve(
      new Response(JSON.stringify({ status: "ready" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
    expect(await screen.findByText("API: online")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByRole("status")).not.toBeInTheDocument());
  });
});
