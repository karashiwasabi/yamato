// File: static/js/dat.js
document.addEventListener("DOMContentLoaded", () => {
  const btn       = document.getElementById("datBtn");
  const input     = document.getElementById("datInput");
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

    // テーブル初期化＋DATヘッダーセット
    thead.innerHTML = `
      <tr>
        <th>卸コード</th><th>日付</th><th>納品／返品</th>
        <th>伝票番号</th><th>行番号</th><th>JANコード</th>
        <th>商品名</th><th>数量</th><th>単価</th>
        <th>小計</th><th>包装薬価</th><th>有効期限</th><th>ロット番号</th>
      </tr>`;
    tbody.innerHTML = "";

    input.click();
  });

  input.addEventListener("change", async () => {
    if (!input.files.length) return;
    indicator.textContent = "DATアップロード中…";

    for (let file of input.files) {
      const form = new FormData();
      form.append("datFileInput[]", file);
      try {
        const res = await fetch("/uploadDat", { method: "POST", body: form });
        const result = await res.json();
        indicator.textContent = `${file.name}: DAT読み込み ${result.DATReadCount}件 | MA0作成 ${result.MA0CreatedCount}件 | 重複 ${result.DuplicateCount}件`;
        // テーブル行追加
        result.DATRecords.forEach(rec => {
          const tr = document.createElement("tr");
          tr.innerHTML = `
            <td>${rec.DatOroshiCode}</td>
            <td>${rec.DatDate}</td>
            <td>${rec.DatDeliveryFlag}</td>
            <td>${rec.DatReceiptNumber}</td>
            <td>${rec.DatLineNumber}</td>
            <td>${rec.DatJanCode}</td>
            <td>${rec.DatProductName}</td>
            <td>${rec.DatQuantity}</td>
            <td>${rec.DatUnitPrice}</td>
            <td>${rec.DatSubtotal}</td>
            <td>${rec.DatPackagingDrugPrice}</td>
            <td>${rec.DatExpiryDate}</td>
            <td>${rec.DatLotNumber}</td>
          `;
          tbody.appendChild(tr);
        });
      } catch (err) {
        console.error(err);
        indicator.textContent = "DATアップロードエラー: " + err.message;
      }
    }
    indicator.textContent += " 完了";
    input.value = "";
  });
});