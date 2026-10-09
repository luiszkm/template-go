import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { applyTheme, readTheme } from "@/lib/theme";
import { App } from "./App";
import { makeQueryClient, makeRouter } from "./router";
import "./index.css";

const root = document.getElementById("root");
if (!root) throw new Error("#root not found");

applyTheme(readTheme());

const queryClient = makeQueryClient();

createRoot(root).render(
  <StrictMode>
    <App router={makeRouter(queryClient)} queryClient={queryClient} />
  </StrictMode>,
);
