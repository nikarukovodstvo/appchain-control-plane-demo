import React, { FormEvent, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import "./styles.css";

type SequencerMode = "single" | "committee";

type ChainConfig = {
  id: string;
  name: string;
  chainId: number;
  rpcUrl: string;
  blockTimeMs: number;
  sequencerMode: SequencerMode;
  createdAt: string;
};

type ChainConfigInput = Omit<ChainConfig, "id" | "createdAt">;

const initial: ChainConfigInput = { name: "orders-l2", chainId: 90101, rpcUrl: "https://rpc.example.test", blockTimeMs: 2000, sequencerMode: "committee" };

function App() {
  const [form, setForm] = useState(initial);
  const [result, setResult] = useState<ChainConfig | null>(null);
  const [status, setStatus] = useState("Ready for a synthetic request");
  const requestKey = useMemo(() => crypto.randomUUID(), []);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setStatus("Creating configuration...");
    try {
      const response = await fetch("/v1/chain-configs", {
        method: "POST",
        headers: { "Content-Type": "application/json", "Idempotency-Key": requestKey },
        body: JSON.stringify(form),
      });
      const body = await response.json();
      if (!response.ok) throw new Error(body.error?.message ?? `HTTP ${response.status}`);
      setResult(body);
      setStatus(response.headers.get("Idempotency-Replayed") === "true" ? "Safe replay: existing record returned" : "Created once");
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Request failed");
    }
  }

  return <main>
    <header><p className="eyebrow">SYNTHETIC TECHNICAL DEMO</p><h1>Appchain Control Plane</h1><p>A small Go + React vertical slice with validation, idempotency, deterministic IDs, and explicit failure states.</p></header>
    <section className="grid">
      <form onSubmit={submit}>
        <h2>Chain configuration</h2>
        <label>Name<input value={form.name} onChange={e => setForm({...form, name:e.target.value})}/></label>
        <label>Chain ID<input type="number" value={form.chainId} onChange={e => setForm({...form, chainId:Number(e.target.value)})}/></label>
        <label>RPC URL<input value={form.rpcUrl} onChange={e => setForm({...form, rpcUrl:e.target.value})}/></label>
        <label>Block time (ms)<input type="number" value={form.blockTimeMs} onChange={e => setForm({...form, blockTimeMs:Number(e.target.value)})}/></label>
        <label>Sequencer<select value={form.sequencerMode} onChange={e => setForm({...form, sequencerMode:e.target.value as SequencerMode})}><option value="single">Single</option><option value="committee">Committee</option></select></label>
        <button>Create / safely replay</button>
      </form>
      <aside><h2>Execution state</h2><p className="status">{status}</p><dl><dt>Request key</dt><dd>{requestKey}</dd><dt>Record ID</dt><dd>{result?.id ?? "Not created"}</dd><dt>Created</dt><dd>{result?.createdAt ?? "-"}</dd></dl><p className="note">Repeat the same request to receive the same record. Reusing the key with different input returns HTTP 409.</p></aside>
    </section>
  </main>;
}

createRoot(document.getElementById("root")!).render(<React.StrictMode><App /></React.StrictMode>);

