// File: static/js/usage.js
document.addEventListener("DOMContentLoaded", () => {
  const btn       = document.getElementById("usageBtn");
  const input     = document.getElementById("usageInput");
  const indicator = document.getElementById("indicator");
  const table     = document.getElementById("outputTable");
  const thead     = table.querySelector("thead");
  const tbody     = table.querySelector("tbody");

  // 初期化
  thead.innerHTML = "";
  tbody.innerHTML = "";

  btn.addEventListener("click", () => {
    // フィルタ部を隠す
    const filterDiv = document.getElementById("aggregateFilter");
    filterDiv.style.display = "none";

    // テーブル初期化＋USAGEヘッダーセット
    thead.innerHTML = `
      <tr>
        <th>日付</th><th>YJコード</th><th>JANコード</th>
        <th>商品名</th><th>数量</th><th>単位コード</th><th>単位名称</th>
      </tr>`;
    tbody.innerHTML = "";

    input.click();
  });

  input.addEventListener("change", async () => {
    if (!input.files.length) return;
    indicator.textContent = "USAGEアップロード中…";

    for (let file of input.files) {
      const form = new FormData();
      form.append("usageFileInput[]", file);
      try {
        const res = await fetch("/uploadUsage", { method: "POST", body: form });
        const result = await res.json();
        indicator.textContent = `${file.name}: USAGE読み込み ${result.TotalRecords}件`;
        // テーブル行追加
        result.USAGERecords.forEach(rec => {
          const tr = document.createElement("tr");
          tr.innerHTML = `
            <td>${rec.usageDate}</td>
            <td>${rec.usageYjCode}</td>
            <td>${rec.usageJanCode}</td>
            <td>${rec.usageProductName}</td>
            <td>${rec.usageAmount}</td>
            <td>${rec.usageUnit}</td>
            <td>${rec.usageUnitName}</td>
          `;
          tbody.appendChild(tr);
        });
      } catch (err) {
        console.error(err);
        indicator.textContent = "USAGEアップロードエラー: " + err.message;
      }
    }
    indicator.textContent += " 完了";
    input.value = "";
  });
});