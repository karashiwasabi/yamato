document.addEventListener("DOMContentLoaded", () => {
  const btn       = document.getElementById("inventoryBtn");
  const input     = document.getElementById("inventoryInput");
  const indicator = document.getElementById("indicator");
  const table     = document.getElementById("outputTable");
  const thead     = table.querySelector("thead");
  const tbody     = table.querySelector("tbody");
  const debug     = document.getElementById("debug");

  // 初期クリア
  thead.innerHTML = "";
  tbody.innerHTML = "";
  debug.textContent = "";

  // 棚卸ボタン押下
  btn.addEventListener("click", () => {
    const filterDiv = document.getElementById("aggregateFilter");
    if (filterDiv) filterDiv.style.display = "none";

    thead.innerHTML = `
      <tr>
        <th>日付</th>
        <th>JANコード</th>
        <th>商品名 (MA0(CSV))</th>
        <th>在庫数</th>
        <th>単位</th>
      </tr>`;
    tbody.innerHTML = "";
    debug.textContent = "";
    input.value = null;
    input.click();
  });

  // ファイル選択後
  input.addEventListener("change", async () => {
    if (!input.files.length) return;
    indicator.textContent = "棚卸CSVアップロード中…";

    const file = input.files[0];
    const form = new FormData();
    form.append("inventoryFile", file);

    try {
      const res = await fetch("/uploadInventory", {
        method: "POST",
        body: form
      });
      // デバッグ：HTTPステータス
      debug.textContent = `HTTP status: ${res.status}\n`;

      if (!res.ok) throw new Error(res.statusText);

      const data = await res.json();
      // デバッグ：返却された JSON 全体
      debug.textContent += JSON.stringify(data, null, 2);

      tbody.innerHTML = "";
      data.inventories.forEach(rec => {
        const displayName = rec.MA0Name
          ? `${rec.MA0Name} (${rec.CSVName})`
          : rec.CSVName;

        const tr = document.createElement("tr");
        tr.innerHTML = `
          <td>${rec.Date}</td>
          <td>${rec.JAN}</td>
          <td>${displayName}</td>
          <td>${rec.Qty}</td>
          <td>${rec.Unit}</td>
        `;
        tbody.appendChild(tr);
      });

      indicator.textContent = `棚卸 ${data.count} 件を取り込みました。`;
    } catch (err) {
      console.error(err);
      indicator.textContent = "棚卸アップロード失敗: " + err.message;
    }
  });
});