<script>
  export let logsKind = "api";
  export let logsLines = [];
  export let onLoadLogs;
</script>

<h3>Logs</h3>
<div style="display:flex;gap:8px;align-items:center;margin-bottom:8px;">
  <select bind:value={logsKind}>
    <option value="api">api</option>
    <option value="audit">audit</option>
  </select>
  <button on:click={onLoadLogs}>Refresh Logs</button>
  <a
    href={`${location.protocol}//${location.hostname}:9090/api/logs/download?kind=${logsKind}`}
    target="_blank"
    rel="noreferrer"
  >
    <button>Download</button>
  </a>
</div>
<pre style="background:#111;color:#d7ffd7;padding:12px;border-radius:8px;max-height:420px;overflow:auto;">{logsLines.join("\n")}</pre>

