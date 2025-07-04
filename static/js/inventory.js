// static/js/inventory.js
document.addEventListener("DOMContentLoaded", () => {
  const btn       = document.getElementById("inventoryBtn");
  const input     = document.getElementById("inventoryInput");
  const indicator = document.getElementById("indicator");
  const table     = document.getElementById("outputTable");
  const thead     = table.querySelector("thead");
  const tbody     = table.querySelector("tbody");
  const debug     = document.getElementById("debug");

  btn.addEventListener("click", () => {
    const filterDiv = document.getElementById("aggregateFilter");
    if (filterDiv) filterDiv.style.display = "none";

    thead.innerHTML =
      `<tr>
         <th>棚卸日</th>
         <th>YJコード</th>
         <th>JANコード</th>
         <th>商品名</th>
         <th>JAN包装数量</th>
         <th>在庫数(包装単位)</th>
         <th>包装単位</th>
         <th>包装単位コード</th>
         <th>在庫数(JAN包装単位)</th>
         <th>JAN包装数量単位</th>
         <th>JAN包装単位コード</th>
       </tr>`;
    tbody.innerHTML = "";
    debug.textContent = "";
    indicator.textContent = "";
    input.value = null;
    input.click();
  });

  input.addEventListener("change", async () => {
    if (!input.files.length) return;
    indicator.textContent = "棚卸CSVアップロード中…";

    const form = new FormData();
    form.append("inventoryFile", input.files[0]);

    try {
      const res = await fetch("/uploadInventory", { method: "POST", body: form });
      debug.textContent = `HTTP status: ${res.status}\n`;
      if (!res.ok) throw new Error(res.statusText);

      const data = await res.json();
      debug.textContent += JSON.stringify(data, null, 2);

      // CSV → Go → DB → Go → JS 経路で受け取ったフィールドをログ
      data.inventories.forEach((rec, idx) => {
      console.log(
        `[inventory.js] #${idx} HousouTaniUnit="${rec.HousouTaniUnit}"`
        + ` InvHousouTaniUnit="${rec.InvHousouTaniUnit}"`
        + ` JanHousouSuuryouUnit="${rec.JanHousouSuuryouUnit}"`
        + ` InvJanHousouSuuryouUnit="${rec.InvJanHousouSuuryouUnit}"`
      );



        // ' を取り除いたあとの文字列
        const trimmedUnit = rec.HousouTaniUnit.replace(/'/g, "");
        const trimmedJan  = rec.JanHousouSuuryouUnit.replace(/'/g, "");
        console.log(
          `[inventory.js] #${idx} trimmed HousouTaniUnit:`, trimmedUnit,
          ` trimmed JanHousouSuuryouUnit:`, trimmedJan
        );
      });

      tbody.innerHTML = "";
      data.inventories.forEach(rec => {
        const tr = document.createElement("tr");
        tr.innerHTML =
          `<td>${rec.InvDate}</td>
           <td>${rec.InvYjCode}</td>
           <td>${rec.InvJanCode}</td>
           <td>${rec.InvProductName}</td>
           <td>${rec.InvJanHousouSuuryouNumber}</td>
           <td>${rec.Qty}</td>
           <td>${rec.HousouTaniUnit}</td>
           <td>${rec.InvHousouTaniUnit}</td>
           <td>${rec.JanQty}</td>
           <td>${rec.JanHousouSuuryouUnit}</td>
           <td>${rec.InvJanHousouSuuryouUnit}</td>`;
        tbody.appendChild(tr);
      });

      indicator.textContent = `棚卸 ${data.count} 件を取り込みました。`;
    }
    catch (err) {
      console.error(err);
      debug.textContent += `\nError: ${err.message}`;
      indicator.textContent = "棚卸アップロード失敗: " + err.message;
    }
  });
});