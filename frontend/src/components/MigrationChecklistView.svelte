<script>
  let { onOpenBlog } = $props();

  const steps = [
    {
      id: "inventory",
      title: "1. Inventory & Current State",
      description: "Document the current SDI router, multiviewer, glue and timing topology with a rough diagram.",
      hint: "Note which sources (cameras, replay, graphics, etc.) are consumed as SDI at which points.",
    },
    {
      id: "timing",
      title: "2. Timing / PTP Preparation",
      description: "Clarify your PTP domain and GMID plan, and verify PTP distribution in the infrastructure.",
      hint: "Make sure it matches the values shown in Dashboard > System / Timing.",
    },
    {
      id: "addressing",
      title: "3. Multicast Address Plan",
      description: "Design your address plan (drives / folders / views) using Address Map & Planner.",
      hint: "Use Planner to define IP blocks corresponding to SDI router output slots.",
    },
    {
      id: "registry",
      title: "4. NMOS Registry & Nodes",
      description: "Verify that the IS-04 registry and your first NMOS nodes are registered correctly.",
      hint: "Use the Topology and Registry Events panels to track node/device/flow counts and events.",
    },
    {
      id: "flows",
      title: "5. Reference Flows & IS-05",
      description: "Define a few critical sources (e.g. studio cameras) as NMOS flows and connect them via IS-05.",
      hint: "Use the Flows + NMOS views to test sender/receiver matching and TAKE operations.",
    },
    {
      id: "operations",
      title: "6. Operational Scenarios",
      description: "Rehearse monitoring, failover and maintenance scenarios (path A/B, maintenance, test signal) in NMOS.",
      hint: "Use Diagnostics / Quick Check and Automation / Checks to automate basic health checks.",
    },
    {
      id: "cutover",
      title: "7. Phased Cutover & Rollback Plan",
      description: "Plan the SDI → NMOS migration in phases and define a rollback option for each step.",
      hint: "Move monitoring and low-risk sources to NMOS first; keep critical paths for the final phases.",
    },
  ];
</script>

<section class="mt-4 space-y-4">
  <div class="flex items-center justify-between gap-3">
    <div>
      <h3 class="text-sm font-semibold text-gray-100">SDI → NMOS Migration Checklist</h3>
      <p class="text-[11px] text-gray-400">
        A condensed version of the blog steps; use it as a project and operations checklist.
      </p>
    </div>
    <button
      class="px-3 py-1.5 rounded-md border border-sky-700 bg-sky-900 text-sky-100 hover:bg-sky-800 text-[11px] font-medium"
      onclick={() => onOpenBlog?.()}
    >
      Open detailed article
    </button>
  </div>

  <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm divide-y divide-gray-800">
    {#each steps as step}
      <div class="p-4 md:p-5">
        <div class="flex items-start justify-between gap-3">
          <div class="space-y-1">
            <h4 class="text-sm font-semibold text-gray-100">{step.title}</h4>
            <p class="text-[11px] text-gray-300">{step.description}</p>
            <p class="text-[11px] text-gray-500 mt-1">
              {step.hint}
            </p>
          </div>
        </div>
      </div>
    {/each}
  </div>
</section>

