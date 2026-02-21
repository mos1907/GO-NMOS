<script>
  import { api } from "../lib/api.js";

  let { token = "" } = $props();

  let baseUrl = $state("");
  let loading = $state(false);
  let applying = $state(false);
  let error = $state("");
  let ioData = $state(null);
  let mapActive = $state(null);

  async function load() {
    const url = baseUrl.trim();
    if (!url) {
      error = "Enter IS-08 base URL (e.g. http://host:port/x-nmos/channel_mapping/v1.0)";
      return;
    }
    error = "";
    loading = true;
    ioData = null;
    mapActive = null;
    try {
      ioData = await api(`/is08/io?base_url=${encodeURIComponent(url)}`, { token });
      mapActive = await api(`/is08/map/active?base_url=${encodeURIComponent(url)}`, { token });
    } catch (e) {
      error = e.message;
      ioData = null;
      mapActive = null;
    } finally {
      loading = false;
    }
  }

  function buildPassThroughRequested() {
    if (!ioData?.inputs || !ioData?.outputs) return null;
    const inputIds = Object.keys(ioData.inputs);
    const firstInput = inputIds[0];
    if (!firstInput) return null;
    const inputChans = ioData.inputs[firstInput].channels?.length ?? 0;
    const requested = {};
    for (const outId of Object.keys(ioData.outputs)) {
      const outChans = ioData.outputs[outId].channels?.length ?? 0;
      const list = [];
      for (let i = 0; i < outChans; i++) {
        if (i < inputChans) list.push({ input: firstInput, channel_index: i });
        else list.push({ mute: true });
      }
      requested[outId] = list;
    }
    return requested;
  }

  function buildStereoRequested() {
    if (!ioData?.inputs || !ioData?.outputs) return null;
    const inputIds = Object.keys(ioData.inputs);
    const firstInput = inputIds[0];
    if (!firstInput) return null;
    const requested = {};
    for (const outId of Object.keys(ioData.outputs)) {
      const outChans = ioData.outputs[outId].channels?.length ?? 0;
      const list = [];
      for (let i = 0; i < outChans; i++) {
        if (i < 2) list.push({ input: firstInput, channel_index: i });
        else list.push({ mute: true });
      }
      requested[outId] = list;
    }
    return requested;
  }

  function build51Requested() {
    if (!ioData?.inputs || !ioData?.outputs) return null;
    const inputIds = Object.keys(ioData.inputs);
    const firstInput = inputIds[0];
    if (!firstInput) return null;
    const requested = {};
    for (const outId of Object.keys(ioData.outputs)) {
      const outChans = ioData.outputs[outId].channels?.length ?? 0;
      const list = [];
      for (let i = 0; i < outChans; i++) {
        if (i < 6) list.push({ input: firstInput, channel_index: i });
        else list.push({ mute: true });
      }
      requested[outId] = list;
    }
    return requested;
  }

  async function applyPreset(presetName) {
    let requested = null;
    if (presetName === "pass_through") requested = buildPassThroughRequested();
    else if (presetName === "stereo") requested = buildStereoRequested();
    else if (presetName === "5.1") requested = build51Requested();
    if (!requested) return;
    await applyActivation({ mode: "activate_immediate", requested });
  }

  async function applyActivation(activation) {
    const url = baseUrl.trim();
    if (!url) {
      error = "Enter IS-08 base URL first.";
      return;
    }
    applying = true;
    error = "";
    try {
      await api("/is08/map/activations", {
        method: "POST",
        token,
        body: { base_url: url, activation },
      });
      mapActive = await api(`/is08/map/active?base_url=${encodeURIComponent(url)}`, { token });
    } catch (e) {
      error = e.message;
    } finally {
      applying = false;
    }
  }
</script>

<section class="mt-4 space-y-4">
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
    <h2 class="text-lg font-semibold text-gray-100 mb-2">Audio Channel Mapping (IS-08)</h2>
    <p class="text-sm text-gray-400 mb-4">
      Connect to an NMOS device that supports IS-08. Enter the Channel Mapping API base URL (spec: <code class="text-gray-300">/x-nmos/channel_mapping/v1.0</code>). For the VirtualTest mock use <code class="text-gray-300">/x-nmos/channelmapping/v1.0</code> (no underscore).
    </p>
    <p class="text-xs text-gray-500 mb-3">
      <strong>Test with mock:</strong> Run <code class="text-gray-400">docker compose up -d</code> in <code class="text-gray-400">virtualtest</code>, then enter <code class="text-gray-400">http://localhost:8084/x-nmos/channelmapping/v1.0</code> and click Load.
    </p>
    <div class="flex flex-wrap items-center gap-2 mb-4">
      <input
        type="text"
        bind:value={baseUrl}
        placeholder="http://localhost:8084/x-nmos/channelmapping/v1.0"
        class="flex-1 min-w-[280px] px-4 py-2 bg-gray-950 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500"
      />
      <button
        class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:opacity-60 text-white text-sm font-medium"
        disabled={loading}
        onclick={load}
      >
        {loading ? "Loading..." : "Load"}
      </button>
    </div>
    {#if error}
      <p class="text-sm text-red-400 mb-4">{error}</p>
    {/if}

    {#if ioData && (ioData.inputs || ioData.outputs)}
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 border-t border-gray-800 pt-4">
        <div>
          <h3 class="text-sm font-medium text-gray-200 mb-2">Inputs</h3>
          <ul class="space-y-2">
            {#each Object.entries(ioData.inputs ?? {}) as [id, inp]}
              <li class="bg-gray-800/50 rounded-lg p-3 border border-gray-700">
                <span class="text-xs font-mono text-gray-400">{id}</span>
                <p class="text-sm text-gray-200">{inp.properties?.name ?? id}</p>
                <p class="text-xs text-gray-500 mt-1">
                  {#each (inp.channels ?? []).map((c) => c.label || "?") as label}
                    <span class="inline-block mr-1 px-1.5 py-0.5 rounded bg-gray-700 text-gray-300">{label}</span>
                  {/each}
                </p>
              </li>
            {/each}
          </ul>
        </div>
        <div>
          <h3 class="text-sm font-medium text-gray-200 mb-2">Outputs</h3>
          <ul class="space-y-2">
            {#each Object.entries(ioData.outputs ?? {}) as [id, out]}
              <li class="bg-gray-800/50 rounded-lg p-3 border border-gray-700">
                <span class="text-xs font-mono text-gray-400">{id}</span>
                <p class="text-sm text-gray-200">{out.properties?.name ?? id}</p>
                <p class="text-xs text-gray-500 mt-1">
                  {#each (out.channels ?? []).map((c) => c.label || "?") as label}
                    <span class="inline-block mr-1 px-1.5 py-0.5 rounded bg-gray-700 text-gray-300">{label}</span>
                  {/each}
                </p>
              </li>
            {/each}
          </ul>
        </div>
      </div>

      <div class="border-t border-gray-800 pt-4 mt-4">
        <h3 class="text-sm font-medium text-gray-200 mb-2">Quick presets</h3>
        <p class="text-xs text-gray-500 mb-2">Apply a channel map from the first input to all outputs (immediate activation).</p>
        <div class="flex flex-wrap gap-2">
          <button
            class="px-3 py-1.5 rounded-md border border-gray-600 bg-gray-800 text-gray-200 hover:bg-gray-700 text-sm disabled:opacity-60"
            disabled={applying}
            onclick={() => applyPreset("pass_through")}
          >
            Pass-through (first input)
          </button>
          <button
            class="px-3 py-1.5 rounded-md border border-gray-600 bg-gray-800 text-gray-200 hover:bg-gray-700 text-sm disabled:opacity-60"
            disabled={applying}
            onclick={() => applyPreset("stereo")}
          >
            Stereo (L, R)
          </button>
          <button
            class="px-3 py-1.5 rounded-md border border-gray-600 bg-gray-800 text-gray-200 hover:bg-gray-700 text-sm disabled:opacity-60"
            disabled={applying}
            onclick={() => applyPreset("5.1")}
          >
            5.1 (L, R, C, LFE, Ls, Rs)
          </button>
          {#if applying}
            <span class="text-xs text-gray-500 self-center">Applying...</span>
          {/if}
        </div>
      </div>

      {#if mapActive && typeof mapActive === "object" && Object.keys(mapActive).length > 0}
        <div class="border-t border-gray-800 pt-4 mt-4">
          <h3 class="text-sm font-medium text-gray-200 mb-2">Active map (current)</h3>
          <pre class="text-xs text-gray-400 bg-gray-950 p-3 rounded overflow-auto max-h-48">{JSON.stringify(mapActive, null, 2)}</pre>
        </div>
      {/if}
    {/if}
  </div>
</section>
