<script>
  export let addressMap = null;
</script>

<h3>Address Map (/24 buckets)</h3>
<p>Total subnets: {addressMap?.total_subnets || 0}</p>
<table style="width:100%;border-collapse:collapse;">
  <thead>
    <tr>
      <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Subnet</th>
      <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Flow Count</th>
      <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">IPs</th>
    </tr>
  </thead>
  <tbody>
    {#each (addressMap?.items || []) as b}
      <tr>
        <td style="border-bottom:1px solid #eee;padding:8px;">{b.subnet}</td>
        <td style="border-bottom:1px solid #eee;padding:8px;">{b.count}</td>
        <td style="border-bottom:1px solid #eee;padding:8px;">{Object.keys(b.flows || {}).slice(0, 6).join(", ")}</td>
      </tr>
    {/each}
  </tbody>
</table>

