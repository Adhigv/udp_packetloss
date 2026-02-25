async function startTest() {
  document.getElementById("status").innerText = "Test running…";

  const res = await fetch("/start", {
    method: "POST",
    headers: {"Content-Type":"application/json"},
    body: JSON.stringify({
      ip: document.getElementById("ip").value,
      pkt_size: +document.getElementById("pkt").value,
      rate_mb: +document.getElementById("rate").value,
      duration: +document.getElementById("dur").value
    })
  });

  const d = await res.json();
  document.getElementById("status").innerText = "Test completed";

  document.getElementById("s_pkts").innerText = d.sender.Packets;
  document.getElementById("s_bytes").innerText = d.sender.Bytes;
  document.getElementById("s_pkt").innerText = d.sender.PktSize;
  document.getElementById("s_rate").innerText = d.sender.TargetRate;
  document.getElementById("s_tp").innerText = d.sender.Throughput.toFixed(2);

  document.getElementById("r_pkts").innerText = d.receiver.Packets;
  document.getElementById("r_bytes").innerText = d.receiver.Bytes;
  document.getElementById("lat").innerText = d.receiver.LatencyMs.toFixed(3);
  document.getElementById("jit").innerText = d.receiver.JitterMs.toFixed(3);

  document.getElementById("loss").innerText = d.absolute_loss;
  document.getElementById("loss_pct").innerText = d.loss_percentage.toFixed(3) + " %";
}
